# Release Process

This project uses [GoReleaser](https://goreleaser.com/) with semantic versioning based on [conventional commits](https://www.conventionalcommits.org/).

## Prerequisites

- Ensure you have GoReleaser installed (or rely on CI to run it)
- Install development tools: `make tools`

## Semantic Versioning

Version numbers are automatically calculated based on your commit messages using [svu](https://github.com/caarlos0/svu):

- `fix:` commits → patch release (v0.1.0 → v0.1.1)
- `feat:` commits → minor release (v0.1.0 → v0.2.0)
- `BREAKING CHANGE:` or `!` → major release (v0.1.0 → v1.0.0)

## Release Steps

### 1. Check the next version

```bash
make next-version
```

This shows what version will be created based on your commits since the last tag.

### 2. Create the release tag

```bash
make create-tag
```

This will:

- Calculate the next semantic version
- Create a Git tag with that version locally
- **Does not push** - you can review the tag first

### 3. Push the tag

Once you're satisfied with the tag:

```bash
make push-tag
```

This pushes the latest tag to origin.

### 4. Publish the release

#### Locally (if not using CI)

```bash
make publish-release
```

#### Via CI (recommended)

The release will be created automatically when the tag is pushed (requires GitHub Actions setup).

## Testing Releases

Test the release process locally without publishing:

```bash
make test-release-locally
```

## First Release

For the first release when no tags exist:

```bash
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
make release
```

## Conventional Commit Examples

```bash
# Patch release
git commit -m "fix: resolve dependency parsing issue"

# Minor release
git commit -m "feat: add support for Python projects"

# Major release
git commit -m "feat!: redesign CLI interface"
```

## Troubleshooting

**"No tags found"**: Create an initial tag manually (see First Release above)

**GoReleaser fails**: Check `.goreleaser.yml` configuration and ensure you have proper Git credentials

**Wrong version calculated**: Review your commit messages and ensure they follow conventional commit format
