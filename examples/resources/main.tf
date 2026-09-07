variable "domain" {
  type = string
}

variable "user_password" {
  type      = string
  sensitive = true
}

resource "purelymail_domain" "main" {
  name = var.domain
}

resource "purelymail_user" "admin" {
  username = "admin@${var.domain}"
  password = var.user_password
}

resource "purelymail_routing_rule" "support" {
  domain_name      = var.domain
  prefix           = false
  match_user       = "support"
  target_addresses = [purelymail_user.admin.username]
}
