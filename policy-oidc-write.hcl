# Allow creating and reading tokens for the "example" role
path "identity/oidc/token/example" {
  capabilities = ["create", "read", "update"]
}

# Optional: Allow reading OIDC keys for token verification
path "identity/oidc/keys" {
  capabilities = ["read"]
}
