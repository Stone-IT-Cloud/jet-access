resource "vault_kv_secret_v2" "busybox_host_1" {
  mount = vault_mount.kv_ssh_hosts.path  # Use the path from the enabled mount
  name  = "ssh/hosts/dev/busybox-host-1" # Path within the KV engine

  # Data payload for the secret
  # Replace with actual details for your BusyBox hosts once SSH is running
  data_json = jsonencode(var.host1_secrets)
}

resource "vault_kv_secret_v2" "busybox_host_2" {
  mount = vault_mount.kv_ssh_hosts.path
  name  = "ssh/hosts/prod/busybox-host-2"

  data_json = jsonencode(var.host2_secrets)
}
