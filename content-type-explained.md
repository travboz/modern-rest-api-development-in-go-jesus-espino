# `Content-Type` header

The `Content-Type` header doesn't change **what the page has access to**. It changes **whether the browser interprets the file as executable HTML at all**.

## Scenario

```text
You have an endpoint that allows you to save a file or text exposed in another API. If you don’t define the content type, the content type could be guessed by the browser. 

So, I can upload an HTML file and the browser will interpret that as HTML, and load it in the browser, with access to things such as cookies, sessions, and other data. 

If you ensure that the content type is set correctly to something such as application/octet-stream or text/plain, the browser won’t interpret it as HTML.
```

Think of it like this:

- `text/html` → browser says *"this is a webpage, let me render it"* — so any `<script>` tags run, JavaScript executes, and it has full access to cookies, localStorage, etc.
- `text/plain` or `application/octet-stream` → browser says *"this is just a file, display it as raw text / download it"* — nothing executes, so the question of access never even comes up.

The **access to cookies and sessions** isn't being taken away — it's that the content **never gets a chance to run** in the first place. There's no JavaScript executing, so there's nothing to reach for those things.

---

This is relevant in the context of **file upload security**. If a user uploads a malicious `.html` file to your server and you serve it back with `Content-Type: text/html`, the browser renders it as a real webpage — and any JavaScript inside it runs with access to the session of whoever views it. That's an XSS attack.

Serving it with `Content-Type: application/octet-stream` instead means the browser just downloads it rather than rendering it — so the malicious script never fires.

So the `Content-Type` isn't a permission system — it's more like the difference between **handing someone a loaded weapon vs. handing them a photo of one**.