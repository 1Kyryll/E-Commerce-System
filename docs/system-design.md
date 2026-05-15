# System Design 

This document present the high overview system design of an application. This system ensures it can handle a real-world scenarios with a great deal of users traffic and a projected workload on the database. System is designed to face and effectively handle such concerns as: eventual consistency, partial failure, concurrency, with explicit and premeditated choices rather than accidental ones. 

## Overview 

The purchase flow of the application takes User from browsing products to purchasing the desired one, while guaranteeing that no item is oversold, no payment is ever lost or made twice, and no order is silently dropped - even with 10,000 customers trying to purchase the same product simultaneously even with only 10 units of it in stock.    

The patterns ensure this application's logic:

1) **Atomic conditional decrement** at the database for inventory contention.
2) **Time-bound reservations** to hold stock during the payment window without blocking other users. 
3) **The transactional outbox pattern** to fan out events to downstream services(email. analytics, recommendations) without the dual-write problem. 

Every other decision listed in the `adr` directory comes out of these three main concepts. 

## High-Level Architecture

```mermaid
flowchart TB
    User((User))
 
    subgraph Frontend
        Web[Next.js<br/>Storefront]
    end
 
    subgraph Backend["Backend Services"]
        GW[API Gateway<br/>REST]
        CAT[Catalog Service]
        CART[Cart Service]
        ORD[Order Service]
    end
 
    subgraph DataLayer["Data Layer"]
        REDIS[(Redis<br/>cache)]
        PDB[(Products DB)]
        CDB[(Carts DB)]
        RDB[(Reservations)]
        ODB[(Orders DB)]
        OUT[(Outbox)]
    end
 
    subgraph Workers["Async Workers"]
        CLEAN[Cleanup Worker]
        PUB[Outbox Publisher]
    end
 
    subgraph External["External"]
        PAY[Payment Provider]
        EMAIL[Email]
        AN[Analytics]
        REC[Recommendations]
    end
 
    User --> Web
    Web -->|REST| GW
    GW -->|gRPC| CAT
    GW -->|gRPC| CART
    GW -->|gRPC| ORD
 
    CAT --> REDIS
    CAT --> PDB
    CART --> CDB
    ORD --> PDB
    ORD --> RDB
    ORD --> ODB
    ORD --> OUT
    ORD -->|HTTPS| PAY
 
    CLEAN --> RDB
    CLEAN --> PDB
    PUB --> OUT
    PUB --> EMAIL
    PUB --> AN
    PUB --> REC
 
    classDef cluster fill:none,stroke:#888,stroke-width:1px,stroke-dasharray: 4 3
    class Frontend,Backend,DataLayer,Workers,External cluster
```

Internal services communicate via gRPC, Client talks to the Gateway service over REST. Services that read from the Catalog use a cache-aside pattern against Redis, everything else hits DB directly thorugh `pgx`/`sqlc`.

## Purchase flow 

```mermaid
sequenceDiagram
    actor U as User
    participant GW as Gateway
    participant CAT as Catalog
    participant CART as Cart Svc
    participant ORD as Order Svc
    participant DB as Postgres
    participant PAY as Payment
 
    U->>GW: GET /products
    GW->>CAT: ListProducts
    CAT->>DB: cache miss → SELECT
    CAT-->>GW: products
    GW-->>U: product list
 
    U->>GW: POST /cart/items
    GW->>CART: AddItem
    CART->>DB: INSERT cart_items
    CART-->>U: 200 OK
 
    U->>GW: POST /checkout (idempotency-key)
    GW->>ORD: PlaceOrder
 
    note right of ORD: BEGIN TRANSACTION
    ORD->>DB: UPDATE products SET inventory=inventory-1<br/>WHERE id=? AND inventory>0 RETURNING id
    DB-->>ORD: row returned ✓
    ORD->>DB: INSERT INTO reservations<br/>(idempotency_key, expires_at=now()+15min)
    note right of ORD: COMMIT
 
    ORD->>PAY: charge(amount, idempotency_key)
    PAY-->>ORD: success
 
    note right of ORD: BEGIN TRANSACTION
    ORD->>DB: INSERT INTO orders ...
    ORD->>DB: INSERT INTO outbox (event=OrderPlaced)
    ORD->>DB: UPDATE reservations SET status='consumed'
    note right of ORD: COMMIT
 
    ORD-->>U: 201 Created (order_id)
 
    note over DB,PAY: Outbox publisher (async, separate process)<br/>polls outbox table → publishes to email/analytics/recs<br/>→ marks events as sent
```

The two `BEGIN TRANSACTION`/`COMMIT` blocks are the two database transactions that anchor the whole flow and ensure that the Order and Inventory change are correctly processed. Everything else - apyment, fan-out can be failed and retried without corrupting state, because the source of truth is always the database.

## Reservation Lifecycle

Reservations prevent from overselling the product during the payment window. It has a defined lifecycle and handles payment failure and expiry of the reservation.

```mermaid
stateDiagram-v2
    [*] --> Active: created with expires_at = now() + 15min<br/>(inventory already decremented)
 
    Active --> Consumed: payment succeeds<br/>order finalized
    Active --> Released: payment fails<br/>OR cleanup worker fires after expiry
 
    Consumed --> [*]: terminal
    Released --> [*]: inventory += 1<br/>terminal
 
    note right of Active
        Atomic with inventory decrement.
        Cannot exist without the inventory
        having been reduced first.
    end note
 
    note right of Released
        Inventory is restored in the
        same transaction that marks
        the reservation as released.
    end note
```

Key invariant here is that **a Reservation in `Active`/`Consumed` state will always result in inventory being decremented.** There is no scenarios where a reservation is created without the inventory change, and vice versa where the inventory cannot be decremented without a reservation.

## Failure Handling 

```mermaid
flowchart TD
    Pay{Payment Result}
 
    Pay -->|Success| Finalize[Finalize order:<br/>INSERT orders + outbox<br/>UPDATE reservation = consumed]
    Pay -->|Hard fail<br/>card declined| Release[Release reservation:<br/>UPDATE status = released<br/>inventory += 1]
    Pay -->|Timeout<br/>unknown| Reconcile[Reconcile via<br/>provider status API]
 
    Reconcile -->|charge succeeded| Finalize
    Reconcile -->|charge did not happen| Release
    Reconcile -->|still unknown| WaitWebhook[Wait for webhook<br/>or retry status check]
 
    WaitWebhook --> Reconcile
 
    Expire[Background: Cleanup Worker]
    Expire -->|every minute| FindExpired[SELECT FROM reservations<br/>WHERE expires_at < now<br/>AND status = 'active']
    FindExpired --> Release
```

Payment **timeout** is a tricky one, because if the network crashes mid-charge we don't know whether money moved. So we need to be sure in order to release a reservation and rollback inventory. 
