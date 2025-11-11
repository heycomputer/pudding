# project upstream doc discovery

**pudd** makes it easy to find and read the _right_ documentation for your project’s dependencies — offline, version-aware and straight from upstream.

---

## Why pudd?

Developers often waste time:

- 🔍 Googling API docs without knowing which version matches their code
- 🌐 Relying on an internet connection just to look up docs
- ⚠️ Reading docs for the _wrong_ version of a dependency

**pudd** fixes that by reading your project’s dependency files, finding the exact versions you use, and fetching their documentation from upstream sources — so you always have the right docs at your fingertips.

---

## Supported project types

- **Elixir** — `mix.exs`
- **JavaScript / TypeScript** — `package.json`
- **Ruby** — `Gemfile`

---

## How it works

1. pudd reads your dependency manifest (e.g. `mix.exs`, `package.json`, or `Gemfile`).
2. It determines the exact version of each dependency.
3. It fetches the corresponding documentation on-demand from upstream sources (e.g. HexDocs, npmjs, rubydoc.info).
4. Docs are cached locally for offline access.

---

## Installation

### Using brew (⚠️ tap is under construction ⚠️)

```bash
# doesn't work yet
brew tap heycomputer/pudd
brew install pudd
```

### Using go install

```bash
go install github.com/heycomputer/pudd@latest
```

This will download, compile, and install the `pudd` binary to your `$GOPATH/bin` directory (or `$GOBIN` if set). Make sure this directory is in your `PATH`.

### Verify installation

```bash
pudd --help
```

---

## Example

```bash
# Discover and open docs for your project's deps
cd your-project
pudd
```