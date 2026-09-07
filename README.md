# Purelymail Terraform Provider

This provider manages Purelymail domains, users, routing rules, and password
reset methods through the Purelymail API.

> **Disclaimer:** This is an unofficial provider. It is not provided,
> supported, or checked by PurelyMail (Add Rabbit LLC).

## Domain setup

Use `purelymail_ownership_code` to retrieve the complete ownership TXT value.
The API returns a value in the form
`purelymail_ownership_proof=<account-specific-value>`. Create that TXT record
and the required Purelymail MX, SPF, DKIM, and DMARC records with your DNS
provider before creating a `purelymail_domain` resource.

DNS may be managed manually or by another Terraform provider in the same
configuration. The domain resource only registers and manages the domain in
Purelymail; it does not create DNS records. During creation it waits for a TXT
record at the domain apex containing `purelymail_ownership_proof=`. The
default wait is 10 minutes and can be changed with the resource `timeouts`
block:

```hcl
resource "purelymail_domain" "main" {
  name = "example.com"

  timeouts = {
    create = "30m"
  }
}
```

## Development

The provider is developed as one module in the multi-provider GitLab
repository. From the repository root:

```sh
make purelymail-test
make purelymail-build
```

Set `PURELYMAIL_API_TOKEN` for live acceptance tests. The API token is never
logged by the provider.

## Limitations

Purelymail app-password endpoints are intentionally not exposed because the
API has no read/list operation and creation is not reliably observable.
Account credit is available as a data source only.
