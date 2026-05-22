# JWTs in Go — `golang-jwt/jwt/v5`

## What is a JWT?

A JSON Web Token (JWT) is a compact, URL-safe token made of three base64-encoded parts separated by dots:

```
header.payload.signature
```

| Part | Contains |
|---|---|
| **Header** | Token type and signing algorithm (e.g. `HS256`) |
| **Payload** | Claims — statements about the subject (user ID, roles, expiry, etc.) |
| **Signature** | Cryptographic proof the token hasn't been tampered with |

---

## What is `KeyFunc`?

`KeyFunc` is a callback function type you provide to the JWT parser. Its job is to **supply the cryptographic key** used to verify the token's signature.

```go
type Keyfunc func(*Token) (any, error)
```

It receives the **partially-parsed** `*Token` (headers and claims are already decoded, but the signature has **not yet been verified**), and must return either the key or an error.

### Why a callback, not just a key?

Because real applications often deal with multiple keys:

- **Key rotation** — new and old keys are both valid during a rollover.
- **Multiple issuers** — each issuer has its own public key.
- **`kid` (Key ID) header** — tokens carry a hint about which key to use.

`KeyFunc` lets you inspect the partially-parsed token and implement whatever key-lookup logic you need before the parser proceeds to signature verification.

### The flow

```
jwt.Parse(tokenString, keyFunc)
    │
    ├── 1. Decode header & claims (base64)
    │
    ├── 2. Call YOUR keyFunc(token) → returns the key
    │
    ├── 3. Use that key to cryptographically verify the signature
    │
    └── 4. Return the verified token (or an error)
```

`KeyFunc` **does not** do the verification itself — it only *supplies the key*. The parser does the actual HMAC/RSA/ECDSA verification internally.

---

## The Critical Algorithm Check

A minimal `KeyFunc` that skips algorithm validation is dangerous:

```go
// ❌ UNSAFE — missing algorithm check
func(token *jwt.Token) (interface{}, error) {
    return secretKey, nil
}
```

Without the algorithm check, an attacker could craft a token claiming `alg: none` (no signature at all) and bypass verification entirely.

```go
// ✅ SAFE — always assert the expected algorithm
func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return secretKey, nil
}
```

The algorithm assertion is the meaningful validation logic that belongs in `KeyFunc`. Signature verification happens after, inside the parser.

---

## Creating a JWT (Best Practice)

```go
package main

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("your-256-bit-secret") // store this securely (env var, secrets manager)

type MyCustomClaims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims       // embeds standard claims: iss, sub, aud, exp, nbf, iat, jti
}

func CreateToken(userID, role string) (string, error) {
    claims := MyCustomClaims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    "your-app-name",
            Subject:   userID,
            Audience:  jwt.ClaimStrings{"your-app-client"},
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // short-lived
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }

    // Create token with chosen signing method
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign and return the complete encoded token as a string
    return token.SignedString(secretKey)
}
```

### Best practice notes

- **Use `RegisteredClaims`** — always include `ExpiresAt`, `IssuedAt`, and `Issuer`.
- **Short expiry** — 15 minutes is a common access token lifetime. Use a separate refresh token for longevity.
- **Embed only what you need** — the payload is base64-encoded, not encrypted. Never put sensitive data (passwords, PII) in claims.
- **Store the secret securely** — use an environment variable or a secrets manager (e.g. AWS Secrets Manager, Vault), never hardcode it.
- **Use asymmetric keys (RS256/ES256) for distributed systems** — if multiple services need to verify tokens, use a private key to sign and distribute only the public key to verifiers.

---

## Verifying / Parsing a JWT (Best Practice)

```go
func VerifyToken(tokenString string) (*MyCustomClaims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &MyCustomClaims{},
        func(token *jwt.Token) (interface{}, error) {
            // 1. Assert the signing algorithm is what we expect
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            // 2. Return the key for the parser to verify the signature
            return secretKey, nil
        },
        // 3. Parser-level options for additional validation
        jwt.WithIssuer("your-app-name"),
        jwt.WithAudience("your-app-client"),
        jwt.WithExpirationRequired(),
        jwt.WithLeeway(5*time.Second), // tolerate minor clock skew between servers
    )
    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    claims, ok := token.Claims.(*MyCustomClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }

    return claims, nil
}
```

### Best practice notes

- **Always check `token.Valid`** — even if `err == nil`, confirm the token is marked valid.
- **Use `ParseWithClaims`** — parse directly into your custom claims struct to avoid an extra type assertion step.
- **Validate `iss` and `aud`** — use `jwt.WithIssuer()` and `jwt.WithAudience()` to prevent tokens from one service being accepted by another.
- **Use `WithExpirationRequired()`** — rejects tokens that were issued without an `exp` claim entirely.
- **Handle errors specifically** — `jwt/v5` returns typed errors you can inspect:

```go
if err != nil {
    switch {
    case errors.Is(err, jwt.ErrTokenExpired):
        // token is valid but expired — prompt re-login or refresh
    case errors.Is(err, jwt.ErrTokenSignatureInvalid):
        // token has been tampered with
    case errors.Is(err, jwt.ErrTokenNotValidYet):
        // token used before its nbf time
    default:
        // malformed or unknown error
    }
}
```

---

## Putting It Together

```go
func main() {
    // Create
    tokenString, err := CreateToken("user-123", "admin")
    if err != nil {
        panic(err)
    }
    fmt.Println("Token:", tokenString)

    // Verify
    claims, err := VerifyToken(tokenString)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Verified! UserID: %s, Role: %s\n", claims.UserID, claims.Role)
}
```

---

## Quick Reference

| Concern | Recommendation |
|---|---|
| Algorithm | Assert `token.Method` in `KeyFunc`; prefer `HS256` (symmetric) or `RS256`/`ES256` (asymmetric) |
| Expiry | Always set `ExpiresAt`; keep access tokens short-lived (15 min) |
| Secret storage | Environment variable or secrets manager; never hardcode |
| Sensitive data | Never put passwords, PII, or secrets in the payload |
| Multi-service | Use asymmetric keys so only the auth server holds the private key |
| Clock skew | Use `jwt.WithLeeway()` to tolerate small time differences between servers |
| Error handling | Use `errors.Is()` against typed JWT errors for precise handling |