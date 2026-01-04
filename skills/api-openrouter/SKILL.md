---
name: api-openrouter
description: Use when calling OpenRouter via the `api` CLI in this repo (chat, models listing). This skill covers provider install, profile secrets/env, inspect, and example usage.
---

# Use the OpenRouter provider

## Quick start (Homebrew install)

```bash
brew install patrickjm/tap/api
```

## Provider setup

1) Install the provider if needed:

```bash
api install ./providers/openrouter.js --name openrouter
```

2) Set the token for the active profile:

```bash
api secret set openrouter token "<OPENROUTER_API_KEY>"
```

3) Optional per-profile env values:

```bash
api env set openrouter OPENROUTER_REFERER "https://example.com"
api env set openrouter OPENROUTER_TITLE "api-cli"
```

4) Inspect available commands:

```bash
api inspect openrouter --json
```

## Common commands

```bash
api openrouter.chat \
  --param model="openai/gpt-4o-mini" \
  --param message="Write a haiku about HTTP" \
  --json
```

```bash
api openrouter.models.list --json
```

## Notes

- Use `-p/--profile <name>` to switch profiles.
- Params are `--param key=value`.
