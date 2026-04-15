# Standalone DOKU payment gateway integration guide

This guide is intentionally repository-agnostic so it can be copied into another application, another repository, or an AI instruction file without depending on this codebase.

It describes the full DOKU hosted checkout lifecycle:

1. initialize a payment order
2. persist a local pending transaction
3. create a hosted checkout session in DOKU
4. redirect the user to the payment URL
5. verify webhook callbacks
6. normalize payment status
7. complete fulfilment once
8. reconcile pending transactions if a callback is missed

## Required application configuration

Your application should provide:

- `DOKU_CLIENT_ID`
- `DOKU_SECRET_KEY`
- `DOKU_ENV` or equivalent environment selector
- `PUBLIC_API_BASE_URL` for server-to-server webhook callbacks
- `PUBLIC_APP_BASE_URL` for browser return URLs
- persistent storage for payment transactions
- persistent storage for the domain entity affected by payment

Use:

- sandbox base URL outside production: `https://api-sandbox.doku.com`
- production base URL for live traffic: `https://api.doku.com`

## Recommended provider responsibilities

Keep DOKU-specific transport details inside a payment provider layer. The rest of the business workflow should call responsibilities equivalent to:

1. `name()` or `get_provider_name()`
2. `create_payment(input)`
3. `get_payment_status(payment_id)`
4. `parse_callback(headers, path, raw_body)`

This keeps checkout creation, signature generation, callback verification, and status mapping out of controllers or handlers.

## Data you should persist locally

Persist enough data to make the integration auditable and idempotent:

- internal transaction ID
- order or invoice number
- provider name
- provider payment ID
- transaction type
- customer or user ID
- amount
- currency
- local status
- payment URL for hosted checkout
- metadata required for fulfilment
- created timestamp
- updated timestamp

Use the local transaction record as the source of truth for whether fulfilment has already happened.

## Status normalization

Normalize DOKU statuses into internal statuses:

- `SUCCESS` or `PAID` -> `succeeded`
- `FAILED`, `EXPIRED`, `CANCEL`, `CANCELLED` -> `failed`
- anything else -> `pending`

## End-to-end flow

### 1. Initialize an internal order

Before calling DOKU:

1. determine what the customer is paying for
2. generate an internal order or invoice number
3. create a local transaction record with `pending` status
4. prepare notification and return URLs

Examples of payable business actions:

- subscription purchase or renewal
- featured listing purchase
- order settlement
- event ticket purchase
- invoice payment

### 2. Create a DOKU hosted checkout

Build a payload that includes:

- order invoice number
- amount
- currency
- language
- payment type for sale flow
- payment due time
- customer ID
- customer name
- customer email
- item description or line items when available
- allowed DOKU payment channels
- webhook override if DOKU expects it in additional information
- browser return/result URL

Then:

1. serialize the request body
2. create a request ID
3. create a request timestamp
4. compute the body digest
5. sign the request with DOKU HMAC rules
6. call DOKU checkout endpoint
7. persist returned payment ID and hosted payment URL
8. return the hosted payment URL to the frontend

### 3. Redirect customer to the DOKU payment URL

The frontend should open or redirect the customer to the hosted payment URL returned from the backend.

### 4. Handle the webhook callback

The callback endpoint should:

1. read raw request headers
2. read the raw request body
3. rebuild the signature using the actual request path and raw body
4. reject the request if the signature is missing or invalid
5. parse the payload
6. locate the local transaction by order or invoice number
7. normalize external payment status
8. update local transaction status
9. run fulfilment only if the payment became successful and was not previously fulfilled

### 5. Fulfil the successful payment once

After a payment succeeds:

1. mark the transaction as completed or succeeded
2. load the affected business entity
3. apply the purchased entitlement
4. save the updated business entity

Examples:

- activate or renew a subscription
- mark a listing as featured for a duration
- mark an order as paid
- grant digital access

### 6. Reconcile pending payments

Callbacks can be delayed or lost. Add a recovery path:

1. find pending DOKU transactions
2. call DOKU status endpoint using saved provider payment ID
3. if DOKU reports success, run the same fulfilment flow used by the webhook
4. if DOKU reports failure, mark the transaction failed

## Python example snippets

These snippets are examples only. Adjust field names, persistence, HTTP framework, and error handling to fit your application.

### Python configuration and constants

```python
import base64
import hashlib
import hmac
import json
import os
import uuid
from dataclasses import dataclass
from datetime import datetime, timezone

import requests


DOKU_SANDBOX_BASE_URL = "https://api-sandbox.doku.com"
DOKU_PRODUCTION_BASE_URL = "https://api.doku.com"
DOKU_CHECKOUT_PATH = "/checkout/v1/payment"


def get_doku_base_url() -> str:
    return (
        DOKU_PRODUCTION_BASE_URL
        if os.getenv("DOKU_ENV") == "production"
        else DOKU_SANDBOX_BASE_URL
    )


def utc_timestamp() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")
```

### Python data structures

```python
@dataclass
class Customer:
    customer_id: str
    name: str
    email: str


@dataclass
class CreatePaymentInput:
    order_id: str
    amount: float
    currency: str
    description: str
    notification_url: str
    return_url: str
    customer: Customer
    auto_redirect: bool = True


@dataclass
class CreatePaymentResult:
    provider: str
    order_id: str
    payment_id: str
    payment_url: str
```

### Python signature helpers

```python
def generate_digest(raw_body: str) -> str:
    digest = hashlib.sha256(raw_body.encode("utf-8")).digest()
    return base64.b64encode(digest).decode("utf-8")


def build_doku_signature(
    client_id: str,
    request_id: str,
    request_timestamp: str,
    request_target: str,
    raw_body: str,
    secret_key: str,
) -> str:
    component = "\n".join(
        [
            f"Client-Id:{client_id}",
            f"Request-Id:{request_id}",
            f"Request-Timestamp:{request_timestamp}",
            f"Request-Target:{request_target}",
            f"Digest:{generate_digest(raw_body)}",
        ]
    )
    signature = hmac.new(
        secret_key.encode("utf-8"),
        component.encode("utf-8"),
        hashlib.sha256,
    ).digest()
    return "HMACSHA256=" + base64.b64encode(signature).decode("utf-8")
```

### Python transaction persistence contract

```python
def save_pending_transaction(transaction: dict) -> None:
    """
    Replace this with real persistence.
    Save:
    - internal transaction ID
    - order ID
    - customer ID
    - amount
    - currency
    - provider
    - status
    - metadata
    """
    raise NotImplementedError


def update_transaction_after_checkout(order_id: str, payment_id: str, payment_url: str) -> None:
    raise NotImplementedError


def find_transaction_by_order_id(order_id: str) -> dict | None:
    raise NotImplementedError


def mark_transaction_status(transaction_id: str, status: str) -> None:
    raise NotImplementedError
```

### Python checkout creation

```python
def create_doku_checkout(payment_input: CreatePaymentInput) -> CreatePaymentResult:
    client_id = os.environ["DOKU_CLIENT_ID"]
    secret_key = os.environ["DOKU_SECRET_KEY"]
    base_url = get_doku_base_url()

    payload = {
        "order": {
            "invoice_number": payment_input.order_id,
            "amount": payment_input.amount,
            "currency": payment_input.currency or "IDR",
            "language": "ID",
            "callback_url": payment_input.return_url,
            "callback_url_result": payment_input.return_url,
            "auto_redirect": payment_input.auto_redirect,
            "line_items": [
                {
                    "id": payment_input.order_id,
                    "name": payment_input.description,
                    "quantity": 1,
                    "price": payment_input.amount,
                }
            ],
        },
        "payment": {
            "type": "SALE",
            "payment_due_date": 60,
        },
        "customer": {
            "id": payment_input.customer.customer_id,
            "name": payment_input.customer.name,
            "email": payment_input.customer.email,
        },
        "payment_methods_type": [
            "VIRTUAL_ACCOUNT_BCA",
            "VIRTUAL_ACCOUNT_BANK_MANDIRI",
            "VIRTUAL_ACCOUNT_BRI",
            "VIRTUAL_ACCOUNT_BNI",
            "EMONEY_OVO",
            "EMONEY_DANA",
            "EMONEY_SHOPEE_PAY",
            "ONLINE_TO_OFFLINE_ALFA",
            "ONLINE_TO_OFFLINE_INDOMARET",
        ],
        "additional_info": {
            "override_notification_url": payment_input.notification_url,
        },
    }

    raw_body = json.dumps(payload, separators=(",", ":"))
    request_id = str(uuid.uuid4())
    request_timestamp = utc_timestamp()
    signature = build_doku_signature(
        client_id=client_id,
        request_id=request_id,
        request_timestamp=request_timestamp,
        request_target=DOKU_CHECKOUT_PATH,
        raw_body=raw_body,
        secret_key=secret_key,
    )

    response = requests.post(
        base_url + DOKU_CHECKOUT_PATH,
        data=raw_body,
        headers={
            "Content-Type": "application/json",
            "Accept": "application/json",
            "Client-Id": client_id,
            "Request-Id": request_id,
            "Request-Timestamp": request_timestamp,
            "Signature": signature,
        },
        timeout=15,
    )
    response.raise_for_status()

    data = response.json()["response"]
    payment_id = data["payment"]["token_id"]
    payment_url = data["payment"]["url"]
    order_id = data["order"].get("invoice_number", payment_input.order_id)

    update_transaction_after_checkout(
        order_id=order_id,
        payment_id=payment_id,
        payment_url=payment_url,
    )

    return CreatePaymentResult(
        provider="doku",
        order_id=order_id,
        payment_id=payment_id,
        payment_url=payment_url,
    )
```

### Python order initialization flow

```python
def start_payment_for_business_action(
    customer_id: str,
    customer_name: str,
    customer_email: str,
    amount: float,
    description: str,
    transaction_type: str,
    metadata: dict,
) -> dict:
    order_id = f"ORDER-{uuid.uuid4().hex[:12].upper()}"
    transaction_id = str(uuid.uuid4())

    save_pending_transaction(
        {
            "transaction_id": transaction_id,
            "order_id": order_id,
            "customer_id": customer_id,
            "type": transaction_type,
            "amount": amount,
            "currency": "IDR",
            "provider": "doku",
            "status": "pending",
            "metadata": metadata,
        }
    )

    result = create_doku_checkout(
        CreatePaymentInput(
            order_id=order_id,
            amount=amount,
            currency="IDR",
            description=description,
            notification_url=f"{os.environ['PUBLIC_API_BASE_URL']}/payments/doku/webhook",
            return_url=f"{os.environ['PUBLIC_APP_BASE_URL']}/payments/result",
            customer=Customer(
                customer_id=customer_id,
                name=customer_name,
                email=customer_email,
            ),
        )
    )

    return {
        "transaction_id": transaction_id,
        "order_id": result.order_id,
        "payment_url": result.payment_url,
    }
```

### Python status mapping

```python
def map_doku_status(status: str) -> str:
    normalized = (status or "").strip().upper()
    if normalized in {"SUCCESS", "PAID"}:
        return "succeeded"
    if normalized in {"FAILED", "EXPIRED", "CANCEL", "CANCELLED"}:
        return "failed"
    return "pending"
```

### Python webhook signature verification

```python
def verify_doku_callback(headers: dict, path: str, raw_body: bytes) -> None:
    client_id = headers.get("Client-Id") or headers.get("client-id") or ""
    request_id = headers.get("Request-Id") or headers.get("request-id") or ""
    request_timestamp = headers.get("Request-Timestamp") or headers.get("request-timestamp") or ""
    signature = headers.get("Signature") or headers.get("signature") or ""

    expected = build_doku_signature(
        client_id=client_id,
        request_id=request_id,
        request_timestamp=request_timestamp,
        request_target=path,
        raw_body=raw_body.decode("utf-8"),
        secret_key=os.environ["DOKU_SECRET_KEY"],
    )

    if not signature or not hmac.compare_digest(signature, expected):
        raise ValueError("invalid DOKU callback signature")
```

### Python webhook handler

```python
def handle_doku_webhook(headers: dict, path: str, raw_body: bytes) -> dict:
    verify_doku_callback(headers=headers, path=path, raw_body=raw_body)

    payload = json.loads(raw_body.decode("utf-8"))
    order_id = payload["order"]["invoice_number"]
    payment_id = payload.get("authorize_id") or payload.get("transaction", {}).get("original_request_id")
    external_status = payload.get("transaction", {}).get("status", "")
    internal_status = map_doku_status(external_status)

    transaction = find_transaction_by_order_id(order_id)
    if not transaction:
        raise LookupError("transaction not found")

    if internal_status == "failed":
        mark_transaction_status(transaction["transaction_id"], "failed")
        return {"status": "failed-recorded"}

    if internal_status == "pending":
        return {"status": "ignored-pending"}

    complete_successful_payment(transaction=transaction, provider_payment_id=payment_id)
    return {"status": "ok"}
```

### Python one-time fulfilment

```python
def complete_successful_payment(transaction: dict, provider_payment_id: str | None = None) -> None:
    if transaction["status"] in {"completed", "succeeded"}:
        return

    mark_transaction_status(transaction["transaction_id"], "completed")

    if provider_payment_id:
        # optionally persist provider_payment_id if it was missing earlier
        pass

    transaction_type = transaction["type"]
    metadata = transaction.get("metadata", {})

    if transaction_type == "subscription":
        apply_subscription_purchase(
            customer_id=transaction["customer_id"],
            metadata=metadata,
        )
    elif transaction_type == "featured_listing":
        apply_listing_feature(
            listing_id=metadata["listing_id"],
            duration_days=metadata["duration_days"],
        )
    elif transaction_type == "order_payment":
        mark_order_paid(
            order_id=metadata["business_order_id"],
        )
```

### Python reconciliation flow

```python
def get_doku_payment_status(payment_id: str) -> str:
    base_url = get_doku_base_url()
    response = requests.get(
        f"{base_url}/checkout/v1/payment/{payment_id}/check-status",
        headers={
            "Accept": "application/json",
            "Authorization": "bearer authnologin",
        },
        timeout=15,
    )
    response.raise_for_status()
    payload = response.json()
    return map_doku_status(payload.get("status", ""))


def reconcile_pending_transaction(transaction: dict) -> None:
    if transaction["provider"] != "doku":
        return
    if transaction["status"] != "pending":
        return
    if not transaction.get("payment_id"):
        return

    status = get_doku_payment_status(transaction["payment_id"])
    if status == "succeeded":
        complete_successful_payment(transaction)
    elif status == "failed":
        mark_transaction_status(transaction["transaction_id"], "failed")
```

## AI instructions for generating DOKU integrations

When asked to integrate DOKU in another application:

- treat this as a provider-based payment integration
- persist a local pending transaction before calling DOKU
- save the provider payment ID and payment URL returned by checkout creation
- require both webhook callback handling and browser return flow
- implement DOKU signature creation for outbound requests
- implement DOKU signature verification for inbound callbacks
- normalize DOKU statuses into internal statuses
- make fulfilment idempotent
- add reconciliation for pending transactions
- keep business fulfilment domain-specific
- if code is requested, prefer complete end-to-end examples over partial fragments
