## 7. Primary keys - UUIDv7, stored as native `uuid`
 
Every row gets a UUIDv7 primary key, stored using Postgres's native `uuid` type (16 bytes, binary). No `bigserial`, no string-form storage, no opaque tokens layered on top.
 
```sql
CREATE TABLE products (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid_v7(),  -- or generated app-side
    -- ...
);
```
 
UUIDv7, standardized in RFC 9562 (2024), is a 128-bit identifier whose high-order 48 bits encode a Unix-epoch millisecond timestamp, with the remaining bits being randomness plus version metadata. The timestamp prefix means newly-generated values are roughly sequential — a UUID generated now sorts after one generated a minute ago — which preserves the B-tree index locality that makes UUIDv4 a notorious performance problem on insert-heavy tables. The randomness keeps them unguessable from the outside.
 
Why this beats the obvious alternatives. Auto-incrementing `bigserial` would be smaller (8 bytes versus 16) and slightly faster on the index, but it leaks information — a competitor reading `/orders/47832` learns roughly how many orders exist — and it forces coordination at the database, which becomes a problem the moment you have more than one writer. UUIDv4 is the version most developers reach for; it's uniformly random, which means inserts hit random spots in the B-tree, fragmenting the index and bloating the cache footprint. On a small table it's invisible; on a 100M-row table it's a measurable cost. UUIDv7 fixes this without sacrificing uniqueness. ULID is functionally equivalent to UUIDv7 — timestamp-prefixed, k-sortable, 128 bits — but encodes to Crockford base32 instead of hex; UUIDv7 won the standardization race and has broader future support across languages. Snowflake-style IDs require a coordination service to issue them; overkill for a single-region deployment, worth revisiting if the system ever shards across regions.
 
On the storage decision: Postgres's native `uuid` type stores 16 bytes binary on disk regardless of how UUIDs are written in queries. Storing them as `text` or `varchar(36)` would more than double the storage cost (37 bytes with the header) and slow comparisons substantially — never do this. The `pgx` driver and `github.com/google/uuid` Go library both work seamlessly with the native type, and `uuid.NewV7()` is the generator of choice on the Go side.
 
One refinement worth mentioning even if not adopted immediately: UUIDv7 values are ugly in URLs (`/products/01928a73-2c64-7000-8c2b-d27c8a39e1f9`). Some teams keep UUIDs as the actual primary key for internal joins and concurrency, but expose a separate human-friendly `slug` or short opaque ID for public URLs (`/products/blue-widget` or `/products/p_8sNqL2v`). This is a presentational layer over the same underlying ID, doesn't change the data-layer decision, and can be added later without migration if it becomes desirable.
 
**Confirmation: UUIDv7 with native `uuid` storage is the right call.** Nothing about the reasoning needs revision. The only serious alternative is `bigserial`, that would be an acceptable option if we had a hard evidence that the extra 8 bytes per row mattered - but we don't.   
 