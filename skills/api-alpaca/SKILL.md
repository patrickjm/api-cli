---
name: api-alpaca
description: Use when interacting with Alpaca via the `api` CLI in this repo (account, orders, positions, watchlists, market data). This skill explains how to install the provider, set secrets/env for a profile, inspect commands, and run Alpaca operations.
---

# Use the Alpaca provider

## Quick start (Homebrew install)

```bash
brew install patrickjm/tap/api
```

## Provider setup

1) Install the provider if needed:

```bash
api install ./providers/alpaca.js --name alpaca
```

2) Set secrets for the active profile:

```bash
api secret set alpaca key "<ALPACA_KEY>"
api secret set alpaca secret "<ALPACA_SECRET>"
```

3) Optional per-profile env values (paper by default):

```bash
api env set alpaca ALPACA_ENDPOINT https://paper-api.alpaca.markets
```

4) Inspect available commands:

```bash
api inspect alpaca --json
```

## Common commands

```bash
api alpaca.account.get --json
api alpaca.orders.list --param status=open --json
api alpaca.orders.create \
  --param symbol=AAPL \
  --param qty=1 \
  --param side=buy \
  --param type=market \
  --param time_in_force=day \
  --json
```

## Notes

- Use `-p/--profile <name>` to switch profiles.
- Params are `--param key=value`; JSON inputs (if any) are passed as raw strings.
