# Purelymail Provider

> **Disclaimer:** This is an unofficial provider. It is not provided,
> supported, or checked by PurelyMail (Add Rabbit LLC).

The provider authenticates with a Purelymail API token:

```hcl
provider "purelymail" {
  api_token = var.purelymail_api_token
}
```

`api_token` can instead be supplied through `PURELYMAIL_API_TOKEN`. The
optional `base_url` is intended for test servers and compatible API endpoints.

## Domain setup workflow

`purelymail_ownership_code` returns the complete TXT record value, including
the `purelymail_ownership_proof=` prefix. Add that value to DNS using your
DNS provider, configure the remaining Purelymail MX, SPF, DKIM, and DMARC
records, and only then create `purelymail_domain`.

DNS can be managed in the same Terraform run with another provider and an
explicit dependency, or it can be configured manually before running this
provider. Before `addDomain` is called, the resource waits for the ownership
TXT record. The default create timeout is 10 minutes and can be configured:

```hcl
resource "purelymail_domain" "main" {
  name = "example.com"

  timeouts = {
    create = "30m"
  }
}
```

This provider does not create DNS records because Purelymail does not expose
their values through its API.

The `purelymail_domain` resource exports a `dns_records` collection containing
the ownership TXT record, MX, SPF, three DKIM CNAME records, and DMARC CNAME
record. Each item has `name`, `type`, `value`, and `priority`; `priority` is
set for the MX record and null for other record types. The ownership value is
retrieved from `getOwnershipCode`, while the remaining values come from
Purelymail's published setup requirements.

Resources:

- `purelymail_domain`
- `purelymail_user`
- `purelymail_routing_rule`
- `purelymail_password_reset_method`

Data sources:

- `purelymail_domains`
- `purelymail_domain`
- `purelymail_ownership_code`
- `purelymail_users`
- `purelymail_user`
- `purelymail_routing_rules`
- `purelymail_password_reset_methods`
- `purelymail_account_credit`

App passwords are intentionally unsupported because the API does not expose a
reliable read/list lifecycle for them.
