resource "kpm_secret" "db_password" {
  path        = "kv/db/prod"
  value       = var.db_password
  type        = "password"
  tags        = ["prod", "db"]
  description = "Production database password"
}

variable "db_password" {
  type      = string
  sensitive = true
}
