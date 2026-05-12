data "kpm_secret" "db_host" {
  path = "kv/db/prod-host"
}

output "db_host" {
  value     = data.kpm_secret.db_host.value
  sensitive = true
}
