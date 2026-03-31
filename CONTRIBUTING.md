# Contributing

## Testing the Plugin Locally

### Prerequisites

- Go 1.26+
- Claude Code CLI

### 1. Build and install the binary

Use the Makefile:

```sh
make install
```

This builds the binary with the correct version from `plugin.json` and copies it to the plugin data directory.

### 2. Load the plugin from your local clone

Use the `--plugin-dir` flag to load the plugin directly without installation:

```sh
claude --plugin-dir /path/to/claude-code-notification/plugin
```

> After making changes to hooks or scripts, run `/reload-plugins` inside Claude Code to pick them up without restarting.

### 3. Run tests

```sh
make unit-test

make integration-test
```

### 4. Test hooks manually

You can simulate a hook event by piping JSON to the binary:

```sh
echo '{"hook_event_name": "Stop", "cwd": "/tmp/test-project"}' | ./notification
```

### 5. Check logs

Hook scripts log to the plugin data directory. Tail it to watch for errors during a Claude session:

```sh
tail -f ~/.claude/plugins/data/claude-code-notification-inline/logs/notification.log
```

## Releasing

See [RELEASING.md](RELEASING.md) for the full release workflow, versioning, and CI pipeline details.
