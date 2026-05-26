# CORS Explained (Like You're 12)

## The Cafeteria Analogy

Imagine you're at school and you want to borrow a cookie from the kid sitting next to you.
That's fine — you're both at the same table. Easy.

But what if you tried to reach across the *entire cafeteria* to grab a cookie from a
stranger's lunchbox without asking? The lunch monitor would stop you immediately.

**That's CORS.**

---

## The Setup

Your browser is the lunch monitor. It has one strict rule:

> "A webpage can freely use stuff from its own website, but if it wants stuff from a
> *different* website, that other website has to explicitly say it's okay."

So if you're on `myforum.com` and your page tries to fetch data from `api.somebank.com`,
the browser stops and asks:

> *"Hey api.somebank.com — did you actually invite myforum.com to talk to you?"*

---

## How the Other Website Says "You're Allowed In"

The other server sends back a special note in its response headers:

```
Access-Control-Allow-Origin: https://myforum.com
```

That's the permission slip. The browser reads it and says *"okay, you're on the guest list"*
and lets the data through.

If there's **no permission slip**, or it lists a different website, the browser **blocks the
response** — even if the server already sent data back.

---

## The Key Insight Most People Miss

> CORS is enforced by the **browser**, not the server.

If you use `curl` or Postman to hit an API directly, CORS doesn't exist — those tools don't
care. CORS is purely a browser safety feature to protect you as a user.

The server already handed over the cookie. The **lunch monitor (browser)** is the one who
won't let it reach your hand without the permission slip.

---

## Why Does It Exist?

Imagine you're logged into your bank in one tab. In another tab, you accidentally visit
`evil-site.com`. Without CORS, that evil site's JavaScript could silently make requests to
your bank *using your login cookies* and steal your money.

CORS prevents that by making sure `evil-site.com` can't read responses from `yourbank.com`
unless the bank explicitly allows it.

---

## The 3-Line Summary

| Situation                                     | Result              |
|-----------------------------------------------|---------------------|
| Same website talking to itself                | ✅ Always fine      |
| Different website, server has permission slip | ✅ Allowed          |
| Different website, no permission slip         | ❌ Browser blocks it |

---

## Preflight Requests

For certain requests (e.g. `PUT`, `DELETE`, or requests with custom headers), the browser
sends a **preflight** — an `OPTIONS` request — *before* the real one, asking:

> "I'm about to send a DELETE request with a custom header. Are you cool with that?"

The server must respond with the appropriate `Access-Control-Allow-*` headers, or the real
request never gets sent.

---

## Key CORS Headers Cheat Sheet

| Header                           | What It Does                                      |
|----------------------------------|---------------------------------------------------|
| `Access-Control-Allow-Origin`    | Which origins are permitted                       |
| `Access-Control-Allow-Methods`   | Which HTTP methods are allowed (GET, POST, etc.)  |
| `Access-Control-Allow-Headers`   | Which request headers are permitted               |
| `Access-Control-Allow-Credentials` | Whether cookies/auth headers can be sent        |
| `Access-Control-Max-Age`         | How long to cache the preflight response (seconds)|

---

## Go Implementation

### Project Structure

```
cors-example/
├── main.go
└── middleware/
    └── cors.go
```

---

### `middleware/cors.go` — Reusable CORS Middleware

```go
package middleware

import (
    "net/http"
    "strings"
)

// CORSConfig holds the configuration for CORS behaviour.
type CORSConfig struct {
    AllowedOrigins   []string // e.g. ["https://myforum.com", "https://myapp.com"]
    AllowedMethods   []string // e.g. ["GET", "POST", "PUT", "DELETE"]
    AllowedHeaders   []string // e.g. ["Content-Type", "Authorization"]
    AllowCredentials bool     // Allow cookies / auth headers
    MaxAge           int      // Preflight cache duration in seconds
}

// DefaultConfig returns a sensible default CORS configuration.
func DefaultConfig() CORSConfig {
    return CORSConfig{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: false,
        MaxAge:           300,
    }
}

// CORS returns an http.Handler middleware that applies CORS headers.
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // Check if the incoming origin is permitted.
            if origin != "" && isAllowedOrigin(cfg.AllowedOrigins, origin) {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            } else if contains(cfg.AllowedOrigins, "*") {
                w.Header().Set("Access-Control-Allow-Origin", "*")
            }

            w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
            w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))

            if cfg.AllowCredentials {
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }

            // Handle preflight (OPTIONS) requests — respond immediately without
            // passing to the actual handler.
            if r.Method == http.MethodOptions {
                w.Header().Set("Access-Control-Max-Age", itoa(cfg.MaxAge))
                w.WriteHeader(http.StatusNoContent) // 204
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// isAllowedOrigin checks if the given origin is in the allowed list.
func isAllowedOrigin(allowed []string, origin string) bool {
    for _, o := range allowed {
        if o == "*" || o == origin {
            return true
        }
    }
    return false
}

func contains(slice []string, val string) bool {
    for _, s := range slice {
        if s == val {
            return true
        }
    }
    return false
}

func itoa(n int) string {
    return strings.TrimSpace(strings.ReplaceAll(fmt.Sprintf("%d", n), " ", ""))
}
```

---

### `main.go` — Wiring It All Together

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "cors-example/middleware"
)

func main() {
    // Configure CORS — only allow requests from these two origins.
    corsConfig := middleware.CORSConfig{
        AllowedOrigins:   []string{"https://myforum.com", "https://myapp.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           600, // Cache preflight for 10 minutes
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/api/data", dataHandler)
    mux.HandleFunc("/api/health", healthHandler)

    // Wrap the entire mux with CORS middleware.
    handler := middleware.CORS(corsConfig)(mux)

    fmt.Println("Server running on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}

// dataHandler returns some example JSON data.
func dataHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Hello from the API!",
        "status":  "ok",
    })
}

// healthHandler is a simple liveness check.
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "healthy")
}
```

---

## What Happens at Runtime

### A Normal Request (allowed origin)

```
Browser → GET /api/data
          Origin: https://myforum.com

Server  → 200 OK
          Access-Control-Allow-Origin: https://myforum.com
          Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
          ...
Browser → ✅ Lets the response through to JavaScript
```

### A Preflight Request (before a PUT with custom headers)

```
Browser → OPTIONS /api/data
          Origin: https://myforum.com
          Access-Control-Request-Method: PUT
          Access-Control-Request-Headers: Authorization

Server  → 204 No Content
          Access-Control-Allow-Origin: https://myforum.com
          Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
          Access-Control-Allow-Headers: Content-Type, Authorization
          Access-Control-Max-Age: 600
Browser → ✅ Clears the real PUT request to proceed
```

### A Blocked Request (disallowed origin)

```
Browser → GET /api/data
          Origin: https://evil-site.com

Server  → (no CORS headers set for this origin)
Browser → ❌ Blocks the response — JavaScript never sees the data
```

---

## Common Gotchas

- **Wildcard + Credentials don't mix.** You cannot use `Access-Control-Allow-Origin: *`
  alongside `Access-Control-Allow-Credentials: true`. You must echo back the specific origin.

- **CORS errors are browser-only.** `curl` and Postman will always succeed — that's expected.
  The restriction lives in the browser, not the server.

- **Your server still processes the request.** Even when CORS blocks a response, the server
  usually already ran the handler. CORS only stops the browser from *reading* the response.

- **Missing OPTIONS handler = broken preflight.** Always make sure your router responds to
  `OPTIONS` requests, otherwise the preflight will 404 and the real request will never fire.

---

## Using the `rs/cors` Library

Rather than writing your own middleware, you can use the popular
[`rs/cors`](https://github.com/rs/cors) package which handles the boilerplate for you.

```go
handler := cors.New(cors.Options{
    AllowedOrigins: []string{"https://mywebapp.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
}).Handler(mux)
```

Breaking down what each field means, back in lunchbox terms:

| Field            | What It Does                                                                 | Lunchbox Analogy                                                  |
|------------------|------------------------------------------------------------------------------|-------------------------------------------------------------------|
| `AllowedOrigins` | Only `https://mywebapp.com` can read responses from this server              | Only the kid from `mywebapp.com` is allowed to take from my lunchbox |
| `AllowedMethods` | Limits which HTTP verbs are permitted                                        | They can look (GET), add (POST), update (PUT), or remove (DELETE) — no other funny business |
| `AllowedHeaders` | Only these headers may be sent along with the request                        | They can only pass me these two specific notes with their request |

### What About Credentials?

Notice `AllowCredentials` is not set here — it defaults to `false`. This means cookies and
`Authorization` headers won't be forwarded even if the origin is allowed.

| Setting                    | Meaning                                                                               |
|----------------------------|-------------------------------------------------------------------------              |
| `AllowCredentials: false`  | Origin can read responses, but cannot send their login badge along                    |
| `AllowCredentials: true`   | Origin can both read responses *and* prove who they are (send cookies / auth headers) |

> ⚠️ If you set `AllowCredentials: true`, you **cannot** use a wildcard `AllowedOrigins: []string{"*"}`.
> You must specify exact origins, otherwise the browser will block it.
