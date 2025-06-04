[![CI](https://github.com/VloRyan/goltmux/actions/workflows/ci.workflow.yml/badge.svg)](https://github.com/VloRyan/goltmux/actions/workflows/ci.workflow.yml)

# GoltMux ⚡

**A blazing-fast, exact-match HTTP router for Go. Built to be strict, safe, and predictable.**

---
![GoltMux Logo](.github/assets/logo.png)

## 🚀 What is GoltMux?

**GoltMux** is a minimal, performance-oriented HTTP request multiplexer for Go that prioritizes **accuracy over
approximation**.

It was built in response to limitations in existing routers like [
`httprouter`](https://github.com/julienschmidt/httprouter) – particularly critical issues such as
[#73](https://github.com/julienschmidt/httprouter/issues/73), where similar paths can **mismatch or wrongly resolve**.

GoltMux **avoids path ambiguity entirely** by applying a **strict best-match-first** algorithm. No fallbacks, no "close
enough" routes – only the *exactly matching route wins*.

---

## ✨ Key Features

- ✅ **Exact path matching** – no unexpected wildcards or sloppy behavior
- ⚡ **Fast** – minimal allocations, optimized path lookup
- 🔒 **Safe** – no path confusion, no unexpected handler execution
- 🧠 **Predictable logic** – every request hits only the clearest, most specific route
- 📦 **Lightweight** – no third-party dependencies
- 🔌 **Idiomatic Go** – designed to drop into any `net/http` server

---

## ❌ httprouter's Problem (Why GoltMux Exists)

Popular routers like `httprouter` fall short in certain edge cases.

For example:

```go
router.GET("/foo/:bar", handler)
router.GET("/foo/bar/baz", otherHandler)
```

A request to /foo/bar/baz may incorrectly match /foo/:bar instead of the more specific /foo/bar/baz. This leads to
unpredictable and unsafe routing behavior.

GoltMux was created to eliminate this class of routing bug.

## 🛠️ Usage

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/vloryan/goltmux"
)

func main() {
	router := goltmux.NewRouter()

	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User profile")
	})

	router.GET("/users/settings", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User settings")
	})

	http.ListenAndServe(":8080", router)
}
```

In GoltMux, a request to `/users/settings` will **never** be misrouted to `/users/:id`.

The router will always prefer the **most specific and exact** match — no ambiguous fallbacks, no surprises.

---

## 🧪 Status

- ✅ Stable core matching engine
- 🚧 Middleware support in progress
- 📚 Full documentation coming soon
- 🧪 Benchmarks and route test suite in development

---

## 📜 License

MIT License © 2025 [Your Name]

---

## 🤝 Contributions

Issues, discussions, and PRs are welcome!

**GoltMux** aims to be:

- 🔍 Strict in matching
- ⚡ Fast in execution
- 🧼 Minimal in design

Help keep it that way.
