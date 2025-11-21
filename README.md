# pudding

Use version-aware fuzzy-search to quickly find the _right_ documentation from your project's dependencies and cache them locally for fast access anytime.

---

## Why pudding?

Developers often waste time:

- 🔍 Googling API docs without knowing which version matches their code
- 🌐 Relying on an internet connection just to look up docs
- ⚠️ Reading docs for the _wrong_ version of a dependency

**pudding** fixes that by reading your project's dependency files, finding the exact versions you use, and fetching their documentation from upstream sources — so you always have the right docs at your fingertips.

---

## Supported project types

- **Elixir** — `mix.exs`
- **Ruby** — `Gemfile`

---

## How it works

1. pudding reads your dependency manifest (e.g. `mix.exs` or `Gemfile`).
2. It determines the exact version of each dependency.
3. It fetches the corresponding documentation on-demand from upstream sources.
4. Docs are cached locally for offline access.

---

## Installation

### Using brew

```bash
brew install heycomputer/tools/pudding
```

### Using go install

```bash
go install github.com/heycomputer/pudding@latest
```

This will download, compile, and install the `pd` binary to your `$GOPATH/bin` directory (or `$GOBIN` if set). Make sure this directory is in your `PATH`.

### Verify installation

```bash
pd --help
```

---

## Roadmap

- [ ] Terminal UI
- [ ] Search within docs

---

## Testing

Run the test suite:

```bash
make test
```

Run tests with verbose output:

```bash
make test-verbose
```

Run tests with coverage:

```bash
make test-coverage
```
