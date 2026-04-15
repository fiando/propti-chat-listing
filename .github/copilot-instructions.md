# DOKU payment gateway integration instructions

When working on a DOKU payment gateway integration, follow the implementation shape already used in this repository.

## Goals

- Keep the integration provider-based so the business flow does not depend on DOKU-specific transport details.
- Keep the instructions generic and language-agnostic.
- Prefer describing required responsibilities, inputs, outputs, persistence, and validation instead of writing framework-specific code.

## Required configuration

Expect the application to provide:

- a DOKU client identifier
- a DOKU secret key
- an environment selector so sandbox and production use different DOKU base URLs
- a public API base URL for webhook callbacks
- a public app base URL for returning the customer to the application after checkout
- persistent storage for payment transactions and the business entity affected by payment

Use sandbox DOKU endpoints outside production and production DOKU endpoints only for live traffic.

## Functions or responsibilities the integration must provide

Design the payment provider layer so it exposes responsibilities equivalent to:

1. identifying the active payment provider
2. creating a hosted checkout payment session
3. checking payment status later for reconciliation
4. validating and parsing callback notifications from DOKU

The business workflow should call those responsibilities instead of embedding DOKU request logic directly inside handlers or controllers.

## Checkout creation flow

When creating a DOKU payment session, the application should:

1. determine the business action being paid for
   - examples in this repository are subscription upgrades and listing promotions
2. generate an internal order or invoice number
3. create a local transaction record with at least:
   - internal transaction ID
   - user or customer ID
   - transaction type
   - amount
   - currency
   - provider name
   - order or invoice number
   - local status set to pending
   - metadata needed later for fulfilment
4. build the DOKU checkout request for the hosted payment endpoint
5. include order details, payment details, and customer details
6. include:
   - server-to-server notification URL
   - browser return/result URL
   - auto-redirect behavior when desired
7. sign the request with the DOKU HMAC signature format using:
   - client identifier
   - request identifier
   - request timestamp
   - request target or path
   - body digest
   - secret key
8. send the request to DOKU
9. persist the returned DOKU payment identifier and hosted payment URL into the local transaction
10. return the hosted payment URL to the frontend or caller

## Checkout payload expectations

The current implementation expects the checkout request to carry:

- order invoice number
- amount
- currency
- language
- payment type set for a sale flow
- payment due time
- customer ID
- customer name
- customer email
- item or description details when available
- allowed DOKU payment channels or payment method types

If a webhook URL override is used, it must be passed in the additional information area expected by DOKU.

## Callback handling flow

The callback endpoint should:

1. receive the raw headers and raw body
2. rebuild the expected DOKU signature using the actual request path and raw body
3. reject the callback if the signature is missing or invalid
4. parse the callback payload to obtain:
   - order or invoice number
   - DOKU payment identifier or authorization identifier
   - transaction status
5. map DOKU statuses into internal statuses:
   - success or paid -> succeeded
   - failed, expired, cancel, or cancelled -> failed
   - everything else -> pending
6. load the local transaction by order or invoice number
7. update the local transaction status
8. run fulfilment logic only once when payment becomes successful

## Fulfilment flow after successful payment

After a payment succeeds, the application should:

1. mark the transaction as completed
2. load the affected business entity
3. apply the purchased entitlement based on transaction type and metadata
4. store the updated business entity

In this repository, fulfilment means either:

- activating or renewing a paid subscription tier
- applying featured or promoted status to a listing for a limited duration

In another application, replace that with the domain-specific effect of a successful payment.

## Reconciliation or recovery flow

The integration should also support a recovery path for missed callbacks:

1. scan pending transactions for DOKU payments
2. call the DOKU payment status endpoint using the saved DOKU payment identifier
3. if DOKU reports success, complete the same fulfilment flow used by callbacks

This prevents stuck pending transactions when a webhook is delayed or lost.

## Persistence expectations

Persist enough data to make the integration idempotent and auditable:

- internal transaction ID
- order or invoice number
- provider payment ID
- provider name
- amount and currency
- local status
- transaction type
- customer or user ID
- timestamps
- metadata needed for fulfilment
- payment URL if checkout is hosted

Use the transaction record as the source of truth for whether fulfilment has already happened.

## AI behavior guidance

When asked to integrate DOKU in another application:

- keep the design provider-oriented
- add explicit transaction persistence before and after checkout creation
- ensure request signing and callback verification are both implemented
- require a webhook endpoint plus a browser return URL
- map external DOKU statuses into internal normalized statuses
- implement idempotent fulfilment
- add a reconciliation path for pending payments
- do not hardcode business fulfilment rules; make them domain-specific
- do not answer with language-specific code unless explicitly requested
