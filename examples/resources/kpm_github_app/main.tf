resource "kpm_github_app" "ci" {
  name            = "ci-deploy-bot"
  app_id          = 12345
  installation_id = 67890
  private_key     = var.gh_app_private_key
}

variable "gh_app_private_key" {
  type      = string
  sensitive = true
}

output "app_fingerprint" {
  value = kpm_github_app.ci.private_key_sha256
}
