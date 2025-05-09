module "vault" {
  source        = "./vault"
  vault_address = var.vault_address
  vault_token   = var.vault_token
  host1_secrets = var.host1_secrets
  host2_secrets = var.host2_secrets
}
