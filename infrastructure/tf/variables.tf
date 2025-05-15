variable "vault_address" {
  description = "The address of the Vault server."
  type        = string

}
variable "vault_token" {
  description = "The token to authenticate with Vault."
  type        = string
  sensitive   = true
}

variable "host1_secrets" {
  description = "Secrets for host1"
  type = object({
    hostname       = string
    ip             = string
    port           = string
    username       = string
    password       = string
    key            = string
    key_passphrase = string
  })
  sensitive = true
}

variable "host2_secrets" {
  description = "Secrets for host1"
  type = object({
    hostname       = string
    ip             = string
    port           = string
    username       = string
    password       = string
    key            = string
    key_passphrase = string
  })
  sensitive = true
}
