# Vault OIDC JWT Setup Guide

This README outlines the steps to configure Vault as an OIDC provider, issue JWT tokens with custom claims, and validate those tokens using the provided `validate.py` script.

---

## **Overview**

The setup allows you to:
1. Configure HashiCorp Vault as an OIDC provider.
2. Define roles in Vault to issue JWTs with specific claims.
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

#### **Create a Userpass Auth Engine***

```bash
vault auth enable userpass
#> Success! Enabled userpass auth method at: userpass/

vault write auth/userpass/users/example password=password
#> Success! Data written to: auth/userpass/users/example
```

Get the accessor ID for the userpass path

```bash
vault auth list -detailed -format json | jq -r '.["userpass/"].accessor'
#> auth_userpass_ae7d9bc4
```

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
  policies="default,oidc-policy"
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
  key="default" \
  template='{"custom_claims": "XYZ"}'
```

Verify the role:
```bash
vault read identity/oidc/role/example
```

---

### **6. Generate a JWT**

Login as the userpass login. 
Use the defined role to generate a JWT

```bash
vault read identity/oidc/token/example
```

The response will include a token. Decode it to inspect claims like `aud`, `sub`, `exp`, `iss`.

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

---

### **8. Obtaining the Audience**

#### Fetch the OIDC Role 

```
❯ vault read identity/oidc/role/example
Key          Value
---          -----
client_id    o7lS6Nqsfb7nGeAq9JXdIux9wO
key          default
template     {"custom_claims": "XYZ"}
ttl          1h
```

The client_id is the audience that is set on the JWT token, so each role can be targeted 
at a service, and the servcie can use said client_id to validate its intended for it.

Custom claims can be used to customise the token even further, and templates can be used
to obtain metadata of the identity (think SCIM groups, etc)


## **Key Configuration Highlights**

### **OIDC Role Configuration**
- **`allowed_redirect_uris`**: Specifies allowed callback URLs. Use `"*"` for testing.
- **`aud`**: Audience in the JWT is the client_id of the oidc role.
- **`templates`**: Adds custom claims, such as `custom_claims`.

### **JWT Validation Highlights**
- The `validate.py` script:
  - Fetches Vault’s JWKS.
  - Verifies the JWT signature.
  - Validates `aud` against the expected audience (`role client id`).
  - Validates custom claims

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


