# Vault OIDC JWT Setup Guide

This README outlines the steps to configure Vault as an OIDC provider, issue JWT tokens with custom claims, and validate those tokens using the provided `validate.py` script.

---

## **Overview**

The setup allows you to:
1. Configure HashiCorp Vault as an OIDC provider.
2. Define roles in Vault to issue JWTs with specific claims.
3. Add a custom audience claim (`custom_aud`) for token validation.
4. Bind an identity/entity to a `userpass` user and create an entity alias.
5. Assign the necessary policy to allow access to the OIDC token endpoint and attach it to the entity.
6. Validate the issued tokens using a Python script (`validate.py`) that:
   - Fetches Vault’s JWKS for public key verification.
   - Decodes and validates the token's claims and signature.

---

## **Steps to Achieve This Setup**

### **1. Configure Vault as an OIDC Provider**
Set up Vault as an OIDC provider:

```bash
vault write identity/oidc/config \
  oidc_discovery_url="http://127.0.0.1:8200/v1/identity/oidc" \
  bound_issuer="http://127.0.0.1:8200/v1/identity/oidc"
```

Verify the configuration:
```bash
vault read identity/oidc/config
```

---

### **2. Bind an Identity/Entity to a Userpass User**
To ensure JWT tokens are linked to an entity, bind the `userpass` user to an identity entity:

#### **Create an Entity**
```bash
vault write identity/entity name="example-entity" policies="default"
```

Retrieve the entity ID:
```bash
vault read identity/entity/name/example-entity
```
Take note of the `entity_id` from the output.

#### **Create an Entity Alias**
Bind the `userpass` user to the entity by creating an alias:

```bash
vault write identity/entity-alias \
  name="example" \
  canonical_id="<entity_id>" \
  mount_accessor="<userpass_accessor>"
```

- Replace `<entity_id>` with the ID retrieved from the previous step.
- Find the `mount_accessor` for the `userpass` authentication method:
  ```bash
  vault auth list -detailed
  ```

---

### **3. Assign Policy to Allow Access to OIDC Token Endpoint**

#### **Create the Policy**
Create a policy to allow the entity to read the OIDC token endpoint:

```hcl
# oidc-policy.hcl
path "identity/oidc/token/example" {
  capabilities = ["read"]
}
```

Save this policy to a file named `oidc-policy.hcl`, then write it to Vault:

```bash
vault policy write oidc-policy oidc-policy.hcl
```

#### **Attach the Policy to the Entity**
Assign the `oidc-policy` to the entity created earlier:

```bash
vault write identity/entity/id/<entity_id> \
  policies="default" \
  identity_policies="oidc-policy"
```

---

### **4. Create a Signing Key**
Vault uses signing keys to sign the JWTs. Create a default signing key:

```bash
vault write identity/oidc/key/default \
  rotation_period=24h \
  verification_ttl=168h
```

List the available keys:
```bash
vault list identity/oidc/key
```

---

### **5. Define an OIDC Role**
Define a role (`example`) that issues JWTs:

```bash
vault write identity/oidc/role/example \
  allowed_redirect_uris="*" \
  ttl="1h" \
  user_claim="sub" \
  bound_audiences="my-service" \
  key="default" \
  templates='{"custom_aud": "my-service"}'
```

Verify the role:
```bash
vault read identity/oidc/role/example
```

---

### **6. Generate a JWT**
Use the defined role to generate a JWT:

```bash
vault read identity/oidc/token/example
```

The response will include a token. Decode it to inspect claims like `custom_aud`, `sub`, `exp`, and `iss`.

---

### **7. Validate the JWT**

#### Fetch Vault’s JWKS Endpoint
Ensure Vault’s JWKS endpoint is accessible:
```bash
curl http://127.0.0.1:8200/v1/identity/oidc/.well-known/keys
```

The endpoint should return the public keys used for JWT verification.

#### Use the `validate.py` Script
Use the provided `validate.py` script to:
1. Fetch the JWKS.
2. Decode and validate the token.
3. Check the `custom_aud` claim.

Run the script:
```bash
python validate.py <your-jwt>
```

Expected output:
- **Valid Token**:
  ```plaintext
  Token is valid!
  Claims: {"custom_aud": "my-service", ...}
  ```
- **Invalid Token**:
  ```plaintext
  Invalid token: custom_aud does not match 'my-service'
  ```

---

## **Key Configuration Highlights**

### **OIDC Role Configuration**
- **`allowed_redirect_uris`**: Specifies allowed callback URLs. Use `"*"` for testing.
- **`bound_audiences`**: Defines the standard `aud` claim. Set to `"my-service"`.
- **`templates`**: Adds custom claims, such as `custom_aud`.

### **JWT Validation Highlights**
- The `validate.py` script:
  - Fetches Vault’s JWKS.
  - Verifies the JWT signature.
  - Validates `custom_aud` against the expected audience (`my-service`).

---

## **Next Steps**
- **Extend Claim Validation**:
  Add additional claim checks (e.g., `sub`, `namespace`) as needed in the `validate.py` script.
- **Deploy to Production**:
  Replace `*` in `allowed_redirect_uris` with actual URLs.
- **Secure Key Rotation**:
  Periodically rotate the signing key using Vault’s key rotation features.

---

For further questions or debugging assistance, feel free to ask!


