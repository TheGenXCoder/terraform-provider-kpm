data "kpm_credential" "openai" {
  type = "llm"
  path = "openai"
}

output "openai_expires_at" {
  value = data.kpm_credential.openai.expires_at
}
