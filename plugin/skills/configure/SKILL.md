---
name: claude-code-notification:configure
description: Create or update claude-code-notification configuration (global or per-project). Use when the user wants to set up notification backends like ntfy, system notifications, sound, or custom scripts.
argument-hint: "[enable ntfy per-project | disable sound global | <leave empty to chat> ]"
---

# Configure claude-code-notification

Help the user create or update their notification configuration.

## Schema

Read the JSON schema at `${CLAUDE_PLUGIN_ROOT}/schema/config.schema.json` before generating any config. The schema defines all valid fields, types, defaults, and descriptions. Always validate your output against it.

## Config locations

- **Global config:** `~/.config/claude-code-notification/config.json` — applies everywhere
- **Per-project config:** `.claude-code-notification.json` in the project root — merges over global at field level (only set fields you want to override)

## Merge behavior

Config loads in 3 layers: defaults -> global -> per-project. Each layer only overrides fields it explicitly sets. Omitting a field means "inherit from the layer below". Setting `"enabled": false` explicitly disables a backend that may be enabled globally.

## Defaults

- `system_notification.enabled` is `true` by default (zero-config desktop notifications)
- All other backends are disabled by default
- `ntfy.server` defaults to `https://ntfy.sh` when not specified

## Instructions

1. Read the schema file at `${CLAUDE_PLUGIN_ROOT}/schema/config.schema.json`
2. Ask the user what they want to configure: which backends to enable, global vs per-project
3. If the config file already exists, read it first and modify — don't overwrite from scratch
4. Write valid JSON matching the schema. Include `"$schema"` pointing to the schema file for editor autocompletion:
   - Global: `"$schema": "https://raw.githubusercontent.com/berk-karaal/claude-code-notification/main/plugin/schema/config.schema.json"`
   - Per-project: same `$schema` value
5. Create parent directories if needed (`mkdir -p ~/.config/claude-code-notification/`)
6. After writing, briefly explain what was configured and how to test it

## Additional Tips

### Sound Backend

For macOS sound backend, prioritize `/System/Library/Sounds/` for built-in sounds. Suggest:

- on_notification: "/System/Library/Sounds/Ping.aiff"
- on_stop: "/System/Library/Sounds/Glass.aiff"
- on_stop_failure: "/System/Library/Sounds/Basso.aiff"

### Ntfy Backend

If the user wants to enable ntfy, use the default server (`https://ntfy.sh`) unless user mentions using custom server. For the topic, if user not specified, use random long topic id to avoid collision. You can use `openssl rand -base64 64 | tr -dc 'a-zA-Z0-9_-' | head -c $((RANDOM % 10 + 55))` bash command to generate topic id. After configuring ntfy, show the used topic and server to the user so they can subscribe it via web app, mobile app, or CLI.

### Custom Script Backend

When configuring a custom script backend, ask the user for the script path and ensure it's an absolute path. Also ask which events they want to trigger the script (on_notification, on_stop, on_stop_failure) and validate that the script exists and is executable.

## Arguments

`$ARGUMENTS` may contain hints like "enable ntfy", "disable sound", "per-project", "global". Use them to skip questions when the intent is clear.
