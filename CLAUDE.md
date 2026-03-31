## Project

**claude-code-notification** — A Claude Code plugin that sends push notifications when Claude finishes a task or needs user input. Hooks into Claude Code's lifecycle events and dispatches notifications through configurable backends (ntfy.sh, system notifications, sound, custom scripts).

### Constraints

- **Platform**: macOS and Linux only
- **Language**: Go — single binary handles all backend logic, config parsing, and OS detection
- **Security**: Notifications must never contain prompts, tool I/O, file paths, session IDs, or environment variables
- **Dependencies**: Go binary is self-contained; system commands (osascript, notify-send, afplay, paplay/aplay) are OS-native
- **Plugin system**: Must conform to Claude Code plugin structure (`plugin/.claude-plugin/plugin.json` + `plugin/hooks/hooks.json`)
- **Binary distribution**: No compiled binaries in git — built and published via GitHub Actions to GitHub Releases

## Structure

```
cmd/claude-code-notification/main.go   # Entry point — reads hook stdin, loads config, dispatches to backends
internal/
  backend/
    backend.go                         # Backend interface + Payload/EventType types
    ntfy/                              # ntfy.sh HTTP POST backend
    sound/                             # OS-native audio playback (afplay/paplay/aplay)
    system/                            # OS-native desktop notifications (osascript/notify-send)
    script/                            # User-supplied custom script execution
  config/                              # Config loading and merging (global + per-project)
  notification/                        # Payload builder from hook event
plugin/                                # Claude Code plugin manifest and hook wiring
test/integration/                      # Integration tests (binary invocation, backend-specific)
```

## Conventions

- Each backend has a `NewBackend(...)` constructor that validates config and returns `(*Backend, error)` or `*Backend`
- Backends return errors from `Send()` — callers decide how to handle them
- Backend packages use package-level `execCommand` / `lookPath` vars for test seam injection
- Tests use `github.com/stretchr/testify/assert`
- Build version injected via `-ldflags "-X main.version=..."`

## Development

See the @CONTRIBUTING.md file.

## Running Tests

```bash
make unit-test               # Runs unit tests
make integration-test        # Runs integration tests
make coverage                # Runs tests with coverage and outputs report
```
