# purelymail_password_reset_method (Resource)

Creates, updates, and deletes an email or phone password-reset method.

## Arguments

- `username` (Required) — Full user email address.
- `type` (Required) — `email` or `phone`.
- `target` (Required) — Email address or phone number.
- `description` (Optional) — Human-readable description.
- `allow_mfa_reset` (Optional) — Allow this method to reset MFA.

The method identity is the combination of `username`, `type`, and `target`.
