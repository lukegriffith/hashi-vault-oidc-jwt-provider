#!/usr/bin/env python
import jwt
import requests
import sys

# Ensure a JWT is passed as an argument
if len(sys.argv) < 2:
    print("Usage: python validate.py <JWT>")
    sys.exit(1)

# Read the JWT from the first argument
token = sys.argv[1]

# Vault Configuration
vault_url = "http://127.0.0.1:8200/v1/identity/oidc/.well-known/keys"
vault_token = "<your-vault-token>"  # Replace with your Vault token

# Fetch the JWKS
headers = {"X-Vault-Token": vault_token}
response = requests.get(vault_url, headers=headers)

if response.status_code != 200:
    raise Exception(
        f"Failed to fetch JWKS: {response.status_code}, {response.text}")

jwks = response.json()

# Ensure "keys" is in the response
if "keys" not in jwks:
    raise Exception(f"Unexpected JWKS response: {jwks}")

# Decode the JWT header to get the 'kid'
header = jwt.get_unverified_header(token)

# Find the public key matching the 'kid'
key = next((k for k in jwks["keys"] if k["kid"] == header["kid"]), None)
if not key:
    raise Exception(f"No matching key found for kid: {header['kid']}")

# Construct the public key
public_key = jwt.algorithms.RSAAlgorithm.from_jwk(key)

# Decode the token without verifying the audience claim
try:
    decoded = jwt.decode(
        token,
        public_key,
        algorithms=["RS256"],
        options={"verify_aud": False},  # Skip default audience validation
        issuer="http://127.0.0.1:8200/v1/identity/oidc"
    )
except jwt.ExpiredSignatureError:
    print("Token has expired!")
    sys.exit(1)
except jwt.InvalidTokenError as e:
    print(f"Invalid token: {e}")
    sys.exit(1)

# Custom audience validation (check 'custom_aud' claim)
expected_audience = "my-service"
if decoded.get("custom_aud") != expected_audience:
    print(f"Invalid token: custom_aud does not match '{expected_audience}'")
    sys.exit(1)

# If all validations pass
print("Token is valid!")
print("Claims:", decoded)
