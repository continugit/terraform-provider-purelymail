# purelymail_routing_rule (Resource)

Creates and manages a Purelymail routing rule. Changes are implemented as a
delete followed by a create because the API has no update operation.

## Arguments

- `domain_name` (Required) — Domain containing the rule.
- `prefix` (Required) — Match a prefix instead of an exact local part.
- `match_user` (Required) — Local part to match.
- `target_addresses` (Required) — Email addresses receiving the forwarded mail.
- `catchall` (Optional) — Whether the rule is a catchall rule.

The resource exports the API-provided numeric `id` and can be imported by ID.
