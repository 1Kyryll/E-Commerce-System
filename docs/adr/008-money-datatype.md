## 8. Money - NUMERIC with explicit currency
 
This is the most error-prone decision in the data layer, so it deserves more space.
 
The simplest *wrong* answer is to use floating-point (`real`, `double precision`, Go's `float64`). Binary floats cannot represent decimal fractions exactly: `0.1 + 0.2` equals `0.30000000000000004`. This is not a Postgres quirk or a Go quirk — it's how IEEE 754 works in every language. Small errors accumulate across additions and rounding, and the bugs they cause (failed reconciliations, off-by-one-cent totals on customer invoices) are invisible until they aren't. Never use float for money. This rule has no exceptions.
 
Among the correct options there are two serious contenders. The first is **integer in the smallest currency unit** — store amounts as integer counts of the smallest unit (cents for EUR, sen for JPY). Postgres's `BIGINT` is 8 bytes and holds values up to ~9.2 × 10¹⁸, so €19.99 is stored as `1999`. Arithmetic is exact, storage is fast, and the integer type works with every language's standard library. The application is responsible for placing the decimal point on display and for being careful with intermediate calculations like proportional tax that produce sub-unit values. Schema would look like:
 
```sql
price_amount   BIGINT  NOT NULL,
price_currency CHAR(3) NOT NULL,  -- ISO 4217: 'EUR', 'USD', 'GBP'
```
 
The second is **NUMERIC** — Postgres's arbitrary-precision exact decimal type. `NUMERIC(18, 4)` gives 18 total digits with 4 after the decimal point, enough for any plausible amount with two extra digits of precision for intermediate calculations. There is no "remember to divide" rule on display, it naturally handles currencies with different subdivisions (Bahraini dinar has 3 decimal places, JPY has 0), and the arithmetic is exact.
 
```sql
price_amount   NUMERIC(18, 4) NOT NULL,
price_currency CHAR(3)        NOT NULL,
```
 
There is also Postgres's built-in `money` type. It exists, but it's locale-dependent (its display behavior changes with the database's `lc_monetary` setting), it doesn't carry currency information of its own, and it has compatibility issues across drivers. Ignore it.
 
**This project uses `NUMERIC(18, 4)`.** The integer-in-cents approach has a specific failure mode — forgetting to divide on display and rendering "1999 EUR" when "€19.99" was intended, or the inverse mistake when accepting input — that NUMERIC eliminates entirely. NUMERIC also handles intermediate calculations gracefully: 23% VAT on €19.99 produces €4.5977, which a NUMERIC column stores faithfully without forcing a premature rounding decision. With integer cents the same operation forces a choice between storing a separately-rounded `tax_amount` (with potential accumulation drift across many line items) or keeping the extra precision somewhere outside the schema. NUMERIC just stores it.
 
Performance is a non-issue at this scale — NUMERIC is slower than BIGINT but still in the millions of ops per second on modern hardware. The integer approach has its own legitimate advocates (tighter storage, faster math, harder to "accidentally do the wrong thing" because there is no decimal point to mis-place), but the display-conversion failure mode is the deciding factor.
 
One rule is non-negotiable regardless of the storage choice: **never store an amount without its currency in the same row**. A `NUMERIC price_amount` column without a paired `currency` column is a future bug waiting for the day a second currency appears.
 
On the Go side, the domain type is small but pays for itself the first time someone tries to add a EUR discount to a USD price:
 
```go
package money
 
import "github.com/shopspring/decimal"
 
type Money struct {
    Amount   decimal.Decimal
    Currency string // ISO 4217
}
 
func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{Amount: m.Amount.Add(other.Amount), Currency: m.Currency}, nil
}
```
 
`github.com/shopspring/decimal` is the standard Go library for arbitrary-precision decimal arithmetic and integrates cleanly with `pgx` via the NUMERIC type. The `Money` struct wraps a `Decimal` plus its currency, so the type system catches EUR-plus-USD bugs at compile time.
 
Multi-currency *display* — showing prices to a German user in EUR, an American in USD — is a separate concern that requires either per-`(product, currency)` price rows or a conversion service running off live rates. Deferred until needed; EUR is the canonical currency for now.
 