# Releasing

This document describes how releases are built, published, and installed for `claude-code-notification`.

## Version source of truth

The version lives in `plugin/.claude-plugin/plugin.json`:

```json
{
  "version": "0.0.1"
}
```

All components read from this file:

| Component             | How it uses the version                                                                      |
| --------------------- | -------------------------------------------------------------------------------------------- |
| `Makefile`            | Reads via `jq` to inject into local builds (`-X main.version=...`)                           |
| `on-session-start.sh` | Reads to determine which release to download and to check if the installed binary is current |
| `on-hook.sh`          | Reads to verify the binary version before dispatching events                                 |
| GoReleaser            | Derives from the git tag (must match `plugin.json`)                                          |

## How to release

1. **Bump the version** in `plugin/.claude-plugin/plugin.json`
2. **Commit** the version change
3. **Tag and push**:
   ```sh
   git tag v0.0.2
   git push origin main --tags
   ```

The tag **must** use the `v` prefix (e.g., `v0.0.2`) and the version after the prefix **must** match `plugin.json`. The release workflow validates this before building -- if they don't match, the workflow fails.

## What the release workflow does

Defined in `.github/workflows/release.yml`, triggered by `v*` tags:

1. **Version check** -- strips the `v` prefix from the tag and compares against `plugin.json`. Fails fast on mismatch.
2. **GoReleaser** builds the Go binary for 4 platform targets:
   - `darwin/amd64` (Intel Mac)
   - `darwin/arm64` (Apple Silicon Mac)
   - `linux/amd64`
   - `linux/arm64`
3. Each binary gets the version injected via `-X main.version={{.Version}}` (GoReleaser strips the `v` prefix automatically, so the binary reports `0.0.2`, not `v0.0.2`).
4. Binaries are packaged into `tar.gz` archives named `notification_{os}_{arch}.tar.gz`.
5. A GitHub Release is created at the tag with all 4 archives and a `checksums.txt` attached.

Configuration: `.goreleaser.yml`

## How installation works (end user)

When a user installs the plugin and starts a Claude session:

1. **`on-session-start.sh`** runs (registered in `plugin/hooks/hooks.json` as a `SessionStart` hook).
2. It reads the expected version from `plugin.json`.
3. If the binary at `${CLAUDE_PLUGIN_DATA}/bin/notification` exists and `--version` matches, it exits early -- nothing to do.
4. Otherwise, it constructs a download URL:
   ```
   https://github.com/berk-karaal/claude-code-notification/releases/download/v{version}/notification_{os}_{arch}.tar.gz
   ```
5. Downloads the archive via `curl`, extracts the `notification` binary, and installs it to the plugin data directory.
6. All errors are logged to `${CLAUDE_PLUGIN_DATA}/logs/notification.log` but never block the Claude session (the script always exits 0).

On subsequent hook events (Stop, StopFailure, Notification), `on-hook.sh` verifies the binary version matches `plugin.json` before dispatching.

## Local development

To build and install locally without a release:

```sh
make install
```

This builds the binary with the version from `plugin.json` and copies it to the plugin data directory. Then load the plugin from your local clone:

```sh
claude --plugin-dir /path/to/claude-code-notification/plugin
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full local development workflow.
