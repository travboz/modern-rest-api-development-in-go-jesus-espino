# Cache-Control: no-cache & ETag Revalidation

## What `Cache-Control: no-cache` Actually Does

Without `no-cache`, the browser might just serve the cached response directly without ever
asking your server — it would only revalidate once `max-age` expires. If you haven't set a
`max-age`, browser behaviour becomes unpredictable and varies between browsers.

`no-cache` removes that ambiguity. It tells the browser:

> "Don't just assume your cached copy is still good, come and ask me first — every time."

So without `no-cache` the browser might think:

> "I have a cached copy with an ETag, I'll just use it."

With `no-cache` it's forced to think:

> "I have a cached copy with an ETag, but I must check with the server before I use it."

---

## The Guaranteed Flow

With `no-cache` in place, this flow is guaranteed on every request:

```
1. Browser has a cached copy with ETag "abc123"
2. no-cache forces it to always revalidate before using that copy
3. Browser sends:  GET /list/1  +  If-None-Match: "abc123"
4. Server computes the hash of the current data
5a. Hash matches  → 304, browser uses its cached copy, no body sent
5b. Hash differs  → 200 + new ETag + new body
```

Without `no-cache`, step 3 might never happen — the browser could silently serve the stale
cached copy instead, and your handler gets zero traffic.

---

## What Each Header Is Responsible For

`no-cache` and `ETag` are doing different jobs:

| Header                    | Job                                                                               |
|-------------------------  |---------------------------------------------------------------------------------- |
| `Cache-Control: no-cache` | Forces the browser to always come back and ask before using its cache             |
| `ETag` / `If-None-Match`  | Makes that check cheap — only re-downloads the body if something actually changed |

`no-cache` ensures the revalidation always happens. `ETag` ensures that revalidation is
efficient. Without `no-cache`, the `ETag` logic might rarely get exercised at all.

---

## The Door Analogy

- `no-cache` → forces the browser to **knock on the door** every time
- `ETag` → lets you answer the door and say *"nothing's changed, go away"* cheaply without
  handing over the full response

Without `no-cache`, the browser doesn't even bother knocking.

---

## Does `no-cache` Force More Handler Code to Run?

Not exactly — it doesn't force more of your handler code to run. It forces the **browser to
make the request in the first place**.

- **Without `no-cache`** — the browser might never send the request at all. It just serves
  the cached copy silently. Your handler gets zero traffic.
- **With `no-cache`** — the browser always sends the request. Your handler runs, computes the
  hash, and then either returns a cheap `304` or a full `200`.

`no-cache` is about **guaranteeing the request reaches your server**. What happens inside the
handler once it arrives is the ETag's job.

---

## The Example

```go
func handleFetchListById(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
        http.Error(w, fmt.Errorf("Invalid id: [%w]", err).Error(), http.StatusBadRequest)
        return
    }

    list, err := repository.GetListByID(id)
    if err != nil {
        switch {
        case errors.Is(err, ErrRecordNotFound):
            http.Error(w, "Shopping list not found", http.StatusNotFound)
        default:
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    data, err := json.Marshal(list)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Forces the browser to always revalidate with the server before
    // using its cached copy — guarantees the If-None-Match check happens.
    w.Header().Set("Cache-Control", "no-cache")

    // Compute a fingerprint of the current response body.
    etag := fmt.Sprintf(`"%x"`, sha256.Sum256(data))

    // If the browser's cached version matches, skip sending the body.
    if match := r.Header.Get("If-None-Match"); match == etag {
        w.WriteHeader(http.StatusNotModified) // 304 — no body needed
        return
    }

    // Content has changed — send the full response with the new ETag.
    w.Header().Set("ETag", etag)
    w.Write(data)
}
```