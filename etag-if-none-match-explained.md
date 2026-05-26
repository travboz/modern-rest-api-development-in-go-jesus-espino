# ETag & If-None-Match

## The Matched Pair

| Header            | Direction        | Meaning                                                         |  Header type |
|-------------------|------------------|-----------------------------------------------------------------| ------------ |
| `ETag`            | Server → Browser | "Here's the fingerprint of what I'm sending you."               | Response     |
| `If-None-Match`   | Browser → Server | "I already have this fingerprint, has anything changed?"        | Request      |

One goes in each direction to make the revalidation round trip work.

---

## ETag First

An `ETag` (Entity Tag) is a **fingerprint of a response**. Your server generates it and sends
it along with the response:

```
GET /api/user/1
→ 200 OK
   ETag: "abc123"
   Body: { "name": "John" }
```

That `"abc123"` is typically a hash of the response content. If the content changes, the hash
changes.

---

## The Problem It Solves

The browser cached that response. 5 minutes later it wants to check if the data is still
fresh. It has two options:

1. Fetch the whole response again — wasteful if nothing changed
2. Ask the server *"is your current version still `"abc123"`?"* — and only download the body
   if something changed

`If-None-Match` is how it asks that question.

---

## The Full Flow

**First request — server sends fingerprint:**
```
GET /api/user/1
← 200 OK
   ETag: "abc123"
   Body: { "name": "John" }
```

**Subsequent request — browser sends the fingerprint back:**
```
GET /api/user/1
   If-None-Match: "abc123"
```

The server computes the current ETag and compares:

**Nothing changed:**
```
← 304 Not Modified
   (no body sent)
```
Browser uses its cached copy. No body transferred.

**Something changed:**
```
← 200 OK
   ETag: "xyz789"
   Body: { "name": "Jane" }
```
Browser discards old cache and stores the new response.

---

## In Go

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    data := []byte(`{"name": "John"}`)

    // Generate a fingerprint of the content
    hash := fmt.Sprintf("%x", md5.Sum(data))

    // Check if the client already has this version
    if r.Header.Get("If-None-Match") == hash {
        w.WriteHeader(http.StatusNotModified) // 304 — no body needed
        return
    }

    // Client doesn't have it or it changed — send the full response
    w.Header().Set("ETag", hash)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(data)
}
```

---

## How It Relates to `no-cache`

`no-cache` and `ETag`/`If-None-Match` are complementary:

| Concept         | Role        | Meaning                                           |
|-----------------|-------------|---------------------------------------------------|
| `no-cache`      | Policy      | Always check with the server before serving from cache |
| `If-None-Match` | Mechanism   | Here's how to check efficiently without re-sending the whole body |

`no-cache` tells the browser *what to do*. `If-None-Match` is *how it does it*.