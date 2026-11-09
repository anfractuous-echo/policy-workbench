# Interview task: policy limits over time

## The application today

This application simulates outgoing payments. All names, recipients and funds are synthetic.

- **Wallets** shows each wallet's currency, balance, reserved funds and available funds.
- **Policies** lets you create, delete and reorder rules.
- **Transfers** lets you create a payment, inspect the matched rule and simulate its next steps.

Each wallet holds USD or EUR. A payment uses the source wallet's currency, with no conversion.

Policies run from top to bottom. The first rule whose conditions all match determines ALLOW or REJECT. If no rule matches, the payment is rejected. The current **Amount less than** condition checks one payment at a time and includes equality. An empty amount field means no amount restriction.

## Existing policies

The following policies are included in the **Northstar Treasury** demo profile, in this order. The examples below use the initial demo data. Use a fresh instance or **Reset demo** to restore it; reset clears any changes in both profiles.

| Order | Policy                    | Sender                        | Receiver                   | Currency | Amount condition | Action |
| ----- | ------------------------- | ----------------------------- | -------------------------- | -------- | ---------------- | ------ |
| 1     | Restricted USD recipients | Any wallet                    | Restricted recipient group | USD      | Any amount       | REJECT |
| 2     | Operations payments       | Daily operations wallet group | Any recipient              | USD      | At most 10,000   | ALLOW  |
| 3     | Treasury fallback         | Any wallet                    | Any recipient              | USD      | Any amount       | ALLOW  |
| 4     | Europe payments           | Europe wallet                 | Suppliers recipient group  | EUR      | Any amount       | ALLOW  |

The **Operations** wallet belongs to **Daily operations**. **Restricted recipient** belongs to the **Restricted** recipient group. **Supplier** is not restricted.

## Example 1: each payment is checked separately

Use initiator **Alex Morgan**, wallet **Operations** and recipient **Supplier**. Create these payments in sequence:

| Step | Payment    | Matched policy      | Result                                                                                          |
| ---- | ---------- | ------------------- | ----------------------------------------------------------------------------------------------- |
| 1    | 6,000 USD  | Operations payments | ALLOW: the payment is at most 10,000.                                                           |
| 2    | 5,000 USD  | Operations payments | ALLOW: this payment is also at most 10,000. The earlier 6,000 is not added.                     |
| 3    | 12,000 USD | Treasury fallback   | ALLOW: the amount does not match Operations payments, so evaluation continues to the next rule. |

Thus, the current 10,000 condition is neither a cumulative limit nor an automatic rejection above 10,000. It decides whether that particular rule matches.

## Example 2: a matching REJECT stops evaluation

Use initiator **Alex Morgan**, wallet **Operations**, recipient **Restricted recipient** and amount **250 USD**.

1. **Restricted USD recipients** matches the currency and recipient group.
2. Its action is REJECT, so the payment is recorded as **REJECTED_BY_POLICY**.
3. The later ALLOW rules are not evaluated. This payment reserves no wallet funds.

## Customer request

Customers want to use policies to control the **total amount of payments over a time period**, in addition to checking individual payments.

You own this feature. The interviewer is your product counterpart: ask questions, clarify the expected behaviour and agree the scope before implementing it. Then make a focused working change, including the UI, and show how you verified it. A short implementation note is enough.

The period feature is intentionally absent from this starter. Run instructions and the current behaviour are documented in [README.md](README.md).
