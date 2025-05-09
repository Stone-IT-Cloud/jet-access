resource "vault_policy" "ssh_hosts_limited_dev_reader" {
  name = "ssh-hosts-limited-dev-reader"

  policy = jsonencode({
    path = {
      "secret/data/ssh/hosts/dev/*" = { # KV v2 data path includes /data/
        capabilities = ["read"]
      },
      "secret/metadata/ssh/hosts/dev/*" = { # KV v2 metadata path for listing
        capabilities = ["list"]
      }
      # Allow listing the base path to see the 'dev' and 'prod' directories
      "secret/metadata/ssh/hosts/*" = {
        capabilities = ["list"]
      }
      "secret/metadata/ssh/*" = {
        capabilities = ["list"]
      }
      # Allow listing the base path to see the 'ssh' or 'hosts' directories
      "secret/metadata/*" = {
        capabilities = ["list"]
      }
    }
  })
}

resource "vault_policy" "ssh_hosts_full_dev_reader" {
  name = "ssh-hosts-full-dev-reader"

  policy = jsonencode({
    path = {
      "secret/data/ssh/hosts/*" = { # KV v2 data path includes /data/
        capabilities = ["read"]
      }
      # Allow listing the base path to see the 'dev' and 'prod' directories
      "secret/metadata/ssh/hosts/*" = {
        capabilities = ["list"]
      }
      "secret/metadata/ssh/*" = {
        capabilities = ["list"]
      }
      # Allow listing the base path to see the 'ssh' or 'hosts' directories
      "secret/metadata/*" = {
        capabilities = ["list"]
      }
    }
  })
}

resource "vault_approle_auth_backend_role" "limited_dev_ssh_access_role" {
  backend        = vault_auth_backend.limited_dev_approle.path
  role_name      = "ssh-access-role"                                # A descriptive name for the role
  token_ttl      = 3600                                             # Token TTL (e.g., 1 hour)
  token_max_ttl  = 86400                                            # Token Max TTL (e.g., 24 hours)
  token_policies = [vault_policy.ssh_hosts_limited_dev_reader.name] # Link the policy to the role
  # Optional: further restrict access based on IP, CIDR, etc.
  # bind_secret_id = true # Requires a SecretID to log in
}

resource "vault_approle_auth_backend_role" "full_dev_ssh_access_role" {
  backend        = vault_auth_backend.full_dev_approle.path
  role_name      = "ssh-access-role"                                                                             # A descriptive name for the role
  token_ttl      = 3600                                                                                          # Token TTL (e.g., 1 hour)
  token_max_ttl  = 86400                                                                                         # Token Max TTL (e.g., 24 hours)
  token_policies = [vault_policy.ssh_hosts_limited_dev_reader.name, vault_policy.ssh_hosts_full_dev_reader.name] # Link the policy to the role
  # Optional: further restrict access based on IP, CIDR, etc.
  # bind_secret_id = true # Requires a SecretID to log in
}
