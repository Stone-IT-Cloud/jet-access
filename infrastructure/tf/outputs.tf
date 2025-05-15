output "vault_address" {
  description = "The address of the Vault server."
  value       = module.vault.vault_address # Assuming you'll add a variable for this
}

# Output the path to the KV secrets engine
output "kv_mount_path" {
  description = "The path where the KV secrets engine is mounted."
  value       = module.vault.kv_mount_path
}

# Output the name of the policy created
output "limited_dev_reader_policy_name" {
  description = "The name of the policy created for reading SSH host secrets."
  value       = module.vault.limited_dev_reader_policy_name
}

output "full_dev_reader_policy_name" {
  description = "The name of the policy created for reading SSH host secrets."
  value       = module.vault.full_dev_reader_policy_name
}

# Output the AppRole RoleID
output "limited_dev_approle_role_id" {
  description = "The RoleID for the SSH access AppRole."
  value       = module.vault.limited_dev_approle_role_id
  sensitive   = true # Mark as sensitive as it's part of the credential
}

output "full_dev_approle_role_id" {
  description = "The RoleID for the SSH access AppRole."
  value       = module.vault.full_dev_approle_role_id
  sensitive   = true # Mark as sensitive as it's part of the credential
}

output "limited_dev_approle_secret_id" {
  description = "A SecretID for the SSH access AppRole (for development/testing only)."
  value       = module.vault.limited_dev_approle_secret_id
  sensitive   = true # Mark as sensitive
}

output "full_dev_approle_secret_id" {
  description = "A SecretID for the SSH access AppRole (for development/testing only)."
  value       = module.vault.full_dev_approle_secret_id
  sensitive   = true # Mark as sensitive
}
