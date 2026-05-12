terraform {
  required_providers {
    kpm = {
      source  = "catalyst9/kpm"
      version = "~> 0.1"
    }
  }
}

# All fields can be set via environment variables:
# KPM_SERVER, KPM_CERT, KPM_KEY, KPM_CA_CERT
provider "kpm" {
  server  = "https://agentkms.local:8443"
  # Use pathexpand() to expand ~ — Terraform does not expand tilde in string values.
  cert    = pathexpand("~/.kpm/certs/client.crt")
  key     = pathexpand("~/.kpm/certs/client.key")
  ca_cert = pathexpand("~/.kpm/certs/ca.crt")
}
