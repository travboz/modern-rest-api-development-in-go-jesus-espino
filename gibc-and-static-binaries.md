# glibc & Static Binaries

## What is `glibc`?

`glibc` (GNU C Library) is the standard C library on Linux. It's the layer that sits between
your programs and the Linux kernel — almost every program running on Linux is talking to
`glibc` whether it knows it or not.

### What It Actually Does

When your code does something like:

```c
printf("hello");
fopen("file.txt");
malloc(1024);
```

Those aren't kernel calls directly — they're `glibc` functions. `glibc` translates them into
the actual low-level system calls the kernel understands (`write`, `open`, `brk` etc.).

Think of it as the **receptionist between your program and the kernel**. Your program speaks a
high-level language, the kernel speaks a very low-level one, and `glibc` translates.

---

### Why You'd Encounter It

The most common place people run into `glibc` explicitly is **Docker**. If you build a Go
binary on a machine that has `glibc` and then try to run it in a minimal container like
`alpine` (which uses `musl` instead of `glibc`), it crashes:

```
/app: /lib/x86_64-linux-gnu/libc.so.6: version 'GLIBC_2.34' not found
```

The binary was compiled expecting `glibc` to be there, but it isn't.

The fix in Go is to build a fully static binary that doesn't depend on `glibc` at all:

```bash
CGO_ENABLED=0 GOOS=linux go build -o app .
```

`CGO_ENABLED=0` tells Go not to link against `glibc`, producing a self-contained binary that
runs anywhere.

---

## Dynamic vs Static Linking

When you build a program, it typically doesn't include everything it needs inside the binary
itself. Instead it says:

> *"I need `printf`, I'll grab that from `glibc` at runtime when I run."*

That's **dynamic linking** — the binary is small, but it has external dependencies it expects
to find on the host machine at runtime. If those dependencies aren't there, or are the wrong
version, it crashes.

A **static binary** is the opposite — at build time, everything the program needs is bundled
directly into the binary itself. No external dependencies. No assumptions about what's
installed on the host.

---

## The Analogy

Think of it like a recipe vs a meal:

- **Dynamic binary** — a recipe card. It tells you what ingredients to go find (`glibc`,
  `musl`, etc.) and assemble at runtime. If the host doesn't have those ingredients, it fails.
- **Static binary** — a fully prepared meal in a box. Everything is already inside. It doesn't
  matter what's in the host's kitchen.

---

## How It Solves the Docker Problem

Dynamic binary flow:
```
Binary runs → "I need glibc" → looks on host → not found → 💥 crash
```

Static binary flow:
```
Binary runs → everything already inside → just works ✅
```

Because the static binary carries everything with it, it doesn't care whether the container is
Alpine, Ubuntu, Debian, or even a completely empty `scratch` container with nothing installed
at all.

---

## In Go

Go makes this particularly easy. By default Go compiles to static binaries, but as soon as you
use `cgo` (which bridges Go and C code) it dynamically links against `glibc`. Disabling it:

```bash
CGO_ENABLED=0 GOOS=linux go build -o app .
```

Tells the Go compiler:

> *"Don't use cgo, don't touch glibc, bundle everything into the binary itself."*

The result is a single self-contained binary you can drop into the most minimal Docker
container possible:

```dockerfile
FROM scratch        # Completely empty — no OS, no glibc, nothing
COPY app /app
ENTRYPOINT ["/app"] # Works perfectly because the binary needs nothing from the host
```

---

## Does `go build` Use Dynamic Linking?

Not by default — Go defaults to **static linking** for pure Go code. Dynamic linking only
kicks in when `cgo` is involved.

`cgo` is enabled by default (`CGO_ENABLED=1`) and certain standard library packages use it
under the hood without you explicitly asking — most notably `net` and `os/user`, which on
Linux call out to `glibc` for DNS resolution and user lookups.

So for `go build my-api.go`:

| Scenario | Result |
|---|---|
| Pure Go code, no `net`/`os/user` | ✅ Static binary — no glibc dependency |
| Uses the `net` package (any HTTP server) | ❌ cgo kicks in → dynamically linked against glibc |

Since virtually every API uses the `net` package, in practice `go build` on an API will
produce a dynamically linked binary on Linux.

---

## cgo Is All-or-Nothing

One `cgo` dependency anywhere in the chain — whether it's your code, a third party library,
or a standard library package like `net` — and the **whole binary** becomes dynamically linked.

You can't have a partially static binary where some parts link against `glibc` and others
don't. The moment `cgo` enters the picture, the entire binary gets dynamically linked.

Which is why `CGO_ENABLED=0` is such a blunt but effective fix — it slams the door on `cgo`
entirely across the whole build, forcing everything back to pure Go implementations and
guaranteeing a static output regardless of what packages you're using.

---

## The Tradeoff

| | Dynamic Binary | Static Binary |
|---|---|---|
| Size | Smaller — dependencies live on the host | Larger — dependencies bundled inside |
| Portability | Depends on host having the right libraries | Runs anywhere, no assumptions |
| Docker | Needs a base image with matching libraries | Can use `scratch` — completely empty |
| Go build | Default when `cgo` is enabled | `CGO_ENABLED=0 go build` |

---

## SQLite and cgo

The most popular Go SQLite driver (`mattn/go-sqlite3`) uses `cgo`. It actually contains the
entire SQLite C source code inside it — `cgo` compiles that C code and bundles it into your
binary at build time.

So your final binary ends up containing:

- Your API code
- The `database/sql` interface
- The entire SQLite C library (compiled in via `cgo`)

No separate SQLite installation is needed on the host or in the container. But because `cgo`
is involved, the binary is **dynamically linked against `glibc`** — so `glibc` must be
present wherever the binary runs.

### What Happens if You Try `CGO_ENABLED=0`?

It won't work. Without `cgo` there's no way to compile the embedded C source code, and
you'll get a build error:

```
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH
```

Or:

```
could not determine kind of name for C.sqlite3
```

### Your Options

**1. Keep dynamic linking and use a glibc-based image** — the simplest path. Your binary
needs `glibc`, so use a base image that ships with it. No code changes needed.

**2. Switch to a pure Go SQLite driver** — `modernc.org/sqlite` is a pure Go port of SQLite
with zero `cgo`. That gets you back to `CGO_ENABLED=0` and a fully static binary:

```go
_ "modernc.org/sqlite"  // instead of mattn/go-sqlite3
"database/sql"
```

**3. Use a smaller glibc-based image** — keep `mattn/go-sqlite3` but swap to something
lighter than `ubuntu:latest` that still ships with `glibc`.

---

## Container Base Images — What's Actually in Them

Different base images ship with different sets of pre-installed dependencies. Choosing the
right one depends on what your binary needs at runtime.

| Base Image | Size | glibc | Shell/Utils | When to Use |
|---|---|---|---|---|
| `scratch` | ~0MB | ❌ | ❌ | Fully static binaries with zero dependencies |
| `alpine` | ~5MB | ❌ (uses musl) | ✅ minimal | Small images — but incompatible with glibc binaries |
| `debian:bookworm-slim` | ~75MB | ✅ | ✅ minimal | Lightweight but glibc-compatible — good default |
| `ubuntu:latest` | ~80MB | ✅ | ✅ full | Full Ubuntu environment, familiar tooling |
| `golang:1.22` | ~800MB | ✅ | ✅ full | Build stage only — never use as a runtime image |

### The Alpine Trap

Alpine uses `musl` instead of `glibc`. They're both C libraries but they're not compatible.
If you build your binary on a machine with `glibc` and then try to run it on Alpine:

```
/app: /lib/x86_64-linux-gnu/libc.so.6: version 'GLIBC_2.34' not found
```

It crashes — even though Alpine has a C library. It just has the wrong one.

### The Right Image for Each Scenario

| Scenario | Right Base Image |
|---|---|
| Pure Go binary, `CGO_ENABLED=0` | `scratch` |
| Go binary with `cgo` / SQLite (`mattn`) | `debian:bookworm-slim` or `ubuntu:latest` |
| Go binary with pure Go SQLite (`modernc`) | `scratch` |
| Anything that needs shell access for debugging | `debian:bookworm-slim` |

### Multi-Stage Builds

A common pattern is to use a full Go image to build, then copy only the binary into a
minimal runtime image — keeping the final image small:

```dockerfile
# Stage 1 — build
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o my-api .

# Stage 2 — runtime (tiny, no Go toolchain)
FROM scratch
COPY --from=builder /app/my-api /my-api
EXPOSE 8080
CMD ["/my-api"]
```

The final image contains only the binary. The Go toolchain, source code, and build
dependencies are left behind in the builder stage and never make it into the final image.