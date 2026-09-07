# purelymail_domain (Resource)

Registers and manages a domain in Purelymail after the ownership TXT record is
available in DNS.

## Example

```hcl
resource "purelymail_domain" "example" {
  name = "example.com"
}
```

## Arguments

- `name` (Required) — Domain name.
- `allow_account_reset` (Optional) — Allow account password resets through this domain.
- `symbolic_subaddressing` (Optional) — Enable symbolic subaddressing.
- `recheck_dns` (Optional) — Request a DNS status recheck.
- `timeouts.create` (Optional) — Maximum time to wait for the ownership TXT record. Defaults to 10 minutes.

## Attributes

The resource exports `dns_records`, a list of records with `name`, `type`,
`value`, and `priority`, plus DNS status attributes `passes_mx`, `passes_spf`,
`passes_dkim`, and `passes_dmarc`.
