# purelymail_user (Resource)

Creates and manages a Purelymail mailbox.

## Arguments

- `username` (Required) — Full email address.
- `password` (Optional, Sensitive) — Required when creating a user; the API cannot read it back.
- `enable_search_indexing` (Optional) — Enable search indexing.
- `enable_password_reset` (Optional) — Enable password reset.
- `require_two_factor_authentication` (Optional) — Require two-factor authentication.

The resource can be imported by full email address.
