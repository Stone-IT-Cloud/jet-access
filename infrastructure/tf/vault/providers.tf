provider "vault" {
  address    = var.vault_address
  token      = var.vault_token
  token_name = "root"
}
