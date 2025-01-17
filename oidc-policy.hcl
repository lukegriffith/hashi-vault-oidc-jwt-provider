# Allow creating and reading tokens for the "example" role
path "identity/oidc/token/example" {
  capabilities = ["read"]
}
