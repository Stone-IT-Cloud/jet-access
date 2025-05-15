resource "vault_mount" "kv_ssh_hosts" {
  path = "secret" # Default path for KV v2
  type = "kv"
  options = {
    version = "2" # Specify KV Version 2
  }
  description = "KV secrets engine for storing SSH host connection details"
}
