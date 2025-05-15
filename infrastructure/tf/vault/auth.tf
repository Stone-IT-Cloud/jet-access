resource "vault_auth_backend" "limited_dev_approle" {
  type = "approle"
  path = "approle/limited-dev" # Default path for approle auth
}

resource "vault_auth_backend" "full_dev_approle" {
  type = "approle"
  path = "approle/full-dev" # Default path for approle auth
}

resource "vault_approle_auth_backend_role_secret_id" "limited_dev_ssh_access_secret_id" {
  backend   = vault_auth_backend.limited_dev_approle.path
  role_name = vault_approle_auth_backend_role.limited_dev_ssh_access_role.role_name
  cidr_list = ["0.0.0.0/0"] # Optional: restrict where this SecretID can be used
}

resource "vault_approle_auth_backend_role_secret_id" "full_dev_ssh_access_secret_id" {
  backend   = vault_auth_backend.full_dev_approle.path
  role_name = vault_approle_auth_backend_role.full_dev_ssh_access_role.role_name
  cidr_list = ["0.0.0.0/0"] # Optional: restrict where this SecretID can be used
}
