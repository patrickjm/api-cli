---
name: api-perplexity
description: Use when querying Perplexity via the `api` CLI in this repo (search, ask, deep). This skill covers provider install, profile secrets/env, inspect, and example usage.
---

# Use the Perplexity provider

## Quick start (Homebrew install)

```bash
brew install patrickjm/tap/api
```

## Provider setup

1) Install the provider if needed:

```bash
api install ./providers/perplexity.js --name perplexity
```

2) Set the token for the active profile:

```bash
api secret set perplexity token "<PERPLEXITY_API_KEY>"
```

3) Inspect available commands:

```bash
api inspect perplexity --json
```

## Common commands

```bash
api perplexity.search --param q="Stock market overview today" --json
api perplexity.ask --param q="Explain RAG briefly" --json
api perplexity.deep --param q="Deep research on battery tech" --json
```

## Notes

- Use `-p/--profile <name>` to switch profiles.
- Params are `--param key=value`.
