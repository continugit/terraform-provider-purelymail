terraform {
  required_providers {
    purelymail = {
      source = "continugit/purelymail"
    }
  }
}

provider "purelymail" {}

data "purelymail_ownership_code" "current" {}

output "ownership_txt_record" {
  value     = data.purelymail_ownership_code.current.code
}

# Add ownership_txt_record as a TXT record with your DNS provider, configure
# the remaining Purelymail DNS records, and then create purelymail_domain.
