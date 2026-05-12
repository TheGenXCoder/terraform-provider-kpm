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
  cert    = "~/.kpm/certs/client.crt"
  key     = "~/.kpm/certs/client.key"
  ca_cert = "~/.kpm/certs/ca.crt"
}
