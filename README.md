# terraform-provider-kpm

Terraform provider for [KPM](https://github.com/TheGenXCoder/kpm) and AgentKMS. Manage secrets and GitHub App registrations as first-class Terraform resources.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- Go >= 1.22 (to build from source)
- A running AgentKMS instance with mTLS certificates

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    kpm = {
      source  = "catalyst9/kpm"
      version = "~> 0.1"
    }
  }
}
```

Run `terraform init` to download the provider from the Terraform Registry.

### Building from Source

```bash
git clone https://github.com/TheGenXCoder/terraform-provider-kpm
cd terraform-provider-kpm
make install
```

This installs the provider binary to `$GOPATH/bin`. To use a locally built provider, add a [dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) to `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "catalyst9/kpm" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

## Provider Configuration

```hcl
provider "kpm" {
  server  = "https://agentkms.local:8443"
  cert    = pathexpand("~/.kpm/certs/client.crt")
  key     = pathexpand("~/.kpm/certs/client.key")
  ca_cert = pathexpand("~/.kpm/certs/ca.crt")
}
```

All four fields can be set via environment variables instead of hardcoding paths:

| Attribute | Environment Variable | Description |
|-----------|----------------------|-------------|
| `server`  | `KPM_SERVER`         | AgentKMS server URL |
| `cert`    | `KPM_CERT`           | Path to mTLS client certificate PEM file |
| `key`     | `KPM_KEY`            | Path to mTLS client key PEM file |
| `ca_cert` | `KPM_CA_CERT`        | Path to CA certificate PEM file |

The mTLS certificates are generated when you run `kpm quickstart`. They live in `~/.kpm/certs/` by default.

## Resources

### `kpm_secret`

Creates, updates, and deletes a secret in AgentKMS.

```hcl
resource "kpm_secret" "db_password" {
  path        = "kv/db/prod"
  value       = var.db_password       # sensitive
  type        = "password"            # optional, defaults to "generic"
  tags        = ["prod", "db"]        # optional
  description = "Production DB password"  # optional
}

variable "db_password" {
  type      = string
  sensitive = true
}
```

**Schema:**

| Attribute     | Type         | Required | Description |
|---------------|--------------|----------|-------------|
| `path`        | string       | yes      | Secret path (e.g. `kv/db/prod`). Changing this forces a new resource. |
| `value`       | string       | yes      | Secret value. Sensitive — redacted from plan output. |
| `type`        | string       | no       | One of: `generic` (default), `api-token`, `ssh-key`, `connection-string`, `jwt`, `password`. |
| `description` | string       | no       | Human-readable description. |
| `tags`        | list(string) | no       | Tags for filtering (e.g. `["prod", "db"]`). |

On `terraform plan`, the provider reads the current value from AgentKMS and surfaces out-of-band changes as a planned update. If the secret is not found, it is removed from state and re-created on the next apply.

---

### `kpm_github_app`

Registers a GitHub App installation in AgentKMS.

```hcl
resource "kpm_github_app" "ci" {
  name            = "ci-deploy-bot"
  app_id          = 12345
  installation_id = 67890
  private_key     = var.gh_app_private_key   # sensitive, write-only
}

variable "gh_app_private_key" {
  type      = string
  sensitive = true
}

output "app_fingerprint" {
  value = kpm_github_app.ci.private_key_sha256
}
```

**Schema:**

| Attribute            | Type   | Required | Description |
|----------------------|--------|----------|-------------|
| `name`               | string | yes      | Unique name for this registration. Changing this forces a new resource. |
| `app_id`             | number | yes      | GitHub App ID (from the App settings page). |
| `installation_id`    | number | yes      | GitHub App Installation ID. |
| `private_key`        | string | yes      | PEM-encoded RSA private key. Sensitive. AgentKMS never returns this after registration. |
| `private_key_sha256` | string | computed | SHA-256 fingerprint of `private_key`. Changing the key in HCL updates this fingerprint and triggers an update. |

**Write-only key:** AgentKMS stores the private key encrypted and never returns it on any read endpoint. Terraform stores the SHA-256 fingerprint (`private_key_sha256`) in state to detect when the key changes. Rotating the key updates the fingerprint and triggers a resource update without requiring `taint`.

## Data Sources

### `kpm_secret`

Reads an existing secret without managing its lifecycle.

```hcl
data "kpm_secret" "db_host" {
  path = "kv/db/prod-host"
}

resource "aws_db_instance" "main" {
  address = data.kpm_secret.db_host.value
}
```

**Schema:**

| Attribute | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `path`    | string | yes      | Secret path. |
| `value`   | string | computed | The secret value. Sensitive. |

---

### `kpm_credential`

Fetches a dynamic or short-lived credential. Refreshed on every `terraform plan` and `terraform apply` — no state is written beyond the current run.

```hcl
data "kpm_credential" "openai" {
  type = "llm"
  path = "openai"
}

# Use data.kpm_credential.openai.value as the API key
output "openai_expires_at" {
  value = data.kpm_credential.openai.expires_at
}
```

**Schema:**

| Attribute    | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `type`       | string | yes      | Credential type. Currently supported: `llm`. |
| `path`       | string | yes      | Credential path (e.g. `openai`, `anthropic`). |
| `value`      | string | computed | The credential value (API key). Sensitive. |
| `expires_at` | string | computed | RFC3339 expiry timestamp. |

Because `expires_at` changes on every fetch, it does not drive spurious diffs in dependent resources.

## Development

```bash
# Build
make build

# Run unit tests
make test

# Run acceptance tests (requires a live AgentKMS instance)
export KPM_SERVER=https://localhost:8443
export KPM_CERT=~/.kpm/certs/client.crt
export KPM_KEY=~/.kpm/certs/client.key
export KPM_CA_CERT=~/.kpm/certs/ca.crt
make testacc

# Install locally for manual testing
make install
```

Unit tests use an in-memory AgentKMS stub server (`internal/testhelpers`) and a mock client (`internal/client/mock.go`) — no live server required.

## Known Limitations

- **Metadata drift** — out-of-band changes to a secret's `description`, `tags`, or `type` via the `kpm` CLI are not detected by `terraform plan`. Only `value` drift is surfaced. This will be addressed in a future release once a `GetMetadata` endpoint is available.
- **Token refresh** — the mTLS session token is acquired once per provider session and not refreshed on expiry. Long-running applies with many resources may encounter auth errors if the token TTL is short.
- **Acceptance tests** — full `TF_ACC` acceptance tests are not yet implemented. They require a live AgentKMS instance and will be added in a follow-on PR.
