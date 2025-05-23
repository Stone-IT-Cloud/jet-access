# Output the Vault address for easy reference
output "vault_address" {
  description = "The address of the Vault server."
  value       = var.vault_address # Assuming you'll add a variable for this
}

# Output the path to the KV secrets engine
output "kv_mount_path" {
  description = "The path where the KV secrets engine is mounted."
  value       = vault_mount.kv_ssh_hosts.path
}

# Output the name of the policy created
output "limited_dev_reader_policy_name" {
  description = "The name of the policy created for reading SSH host secrets."
  value       = vault_policy.ssh_hosts_limited_dev_reader.name
}

output "full_dev_reader_policy_name" {
  description = "The name of the policy created for reading SSH host secrets."
  value       = vault_policy.ssh_hosts_full_dev_reader.name
}

# Output the AppRole RoleID
output "limited_dev_approle_role_id" {
  description = "The RoleID for the SSH access AppRole."
  value       = vault_approle_auth_backend_role.limited_dev_ssh_access_role.role_id
  sensitive   = true # Mark as sensitive as it's part of the credential
}

output "full_dev_approle_role_id" {
  description = "The RoleID for the SSH access AppRole."
  value       = vault_approle_auth_backend_role.full_dev_ssh_access_role.role_id
  sensitive   = true # Mark as sensitive as it's part of the credential
}

output "limited_dev_approle_secret_id" {
  description = "A SecretID for the SSH access AppRole (for development/testing only)."
  value       = vault_approle_auth_backend_role_secret_id.limited_dev_ssh_access_secret_id.secret_id
  sensitive   = true # Mark as sensitive
}

output "full_dev_approle_secret_id" {
  description = "A SecretID for the SSH access AppRole (for development/testing only)."
  value       = vault_approle_auth_backend_role_secret_id.full_dev_ssh_access_secret_id.secret_id
  sensitive   = true # Mark as sensitive
}
