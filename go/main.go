package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	jwksURL   = "http://127.0.0.1:8200/v1/identity/oidc/.well-known/keys"
	publicKey *rsa.PublicKey
)

func main() {
	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Load JWKS public key
	err := loadJWKS()
	if err != nil {
		e.Logger.Fatalf("Failed to load JWKS: %v", err)
	}

	// Define Routes
	e.GET("/", helloHandler)
	e.GET("/secure", jwtMiddleware(secureHandler))

	// Start Server
	e.Logger.Fatal(e.Start(":8080"))
}

// helloHandler is the public "Hello World" endpoint
func helloHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

// jwtMiddleware validates the JWT before passing to the handler
func jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing Authorization header"})
		}

		// Extract token from the "Bearer" scheme
		token := authHeader[len("Bearer "):]
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		// Parse and validate the JWT
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is RSA
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}

		// Add claims to context for downstream use
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
			c.Set("claims", claims)
			return next(c)
		}

		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
	}
}

// secureHandler is a protected endpoint
func secureHandler(c echo.Context) error {
	claims := c.Get("claims").(jwt.MapClaims)
	user := claims["sub"] // 'sub' is a standard OIDC claim
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Secure endpoint accessed",
		"user":    user,
		"claims":  claims,
	})
}

// loadJWKS fetches and parses the public key from Vault's JWKS endpoint
func loadJWKS() error {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read JWKS response: %v", err)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"` // Modulus
			E   string `json:"e"` // Exponent
		} `json:"keys"`
	}

	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("failed to unmarshal JWKS: %v", err)
	}

	// Use the first key (you can match `kid` if needed)
	if len(jwks.Keys) == 0 {
		return fmt.Errorf("no keys found in JWKS")
	}

	// Decode modulus (N)
	modulus, err := base64.RawURLEncoding.DecodeString(jwks.Keys[0].N)
	if err != nil {
		return fmt.Errorf("failed to decode modulus: %v", err)
	}

	// Decode exponent (E)
	exponent, err := base64.RawURLEncoding.DecodeString(jwks.Keys[0].E)
	if err != nil {
		return fmt.Errorf("failed to decode exponent: %v", err)
	}

	// Convert exponent to integer
	exponentInt := int(new(big.Int).SetBytes(exponent).Uint64())

	// Convert modulus and exponent to RSA public key
	publicKey = &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulus),
		E: exponentInt,
	}

	return nil
}
