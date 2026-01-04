---
name: api-replicate
description: Use when calling Replicate via the `api` CLI in this repo (models discovery, predictions create/get/wait/cancel). This skill covers provider install, profile secrets/env, inspect, and example usage.
---

# Use the Replicate provider

## Quick start (Homebrew install)

```bash
brew install patrickjm/tap/api
```

## Provider setup

1) Install the provider if needed:

```bash
api install ./providers/replicate.js --name replicate
```

2) Set the token for the active profile:

```bash
api secret set replicate token "<REPLICATE_API_KEY>"
```

3) Inspect available commands:

```bash
api inspect replicate --json
```

## Common commands

```bash
api replicate.models.search --param q="sdxl" --json
api replicate.models.get --param owner="stability-ai" --param name="stable-diffusion-xl" --json
```

Create a prediction (input must be JSON string):

```bash
api replicate.predictions.create \
  --param version="<MODEL_VERSION_ID>" \
  --param input='{"prompt":"A cat astronaut"}' \
  --json
```

Wait for completion:

```bash
api replicate.predictions.wait --param id="<PREDICTION_ID>" --json
```

## Notes

- Use `-p/--profile <name>` to switch profiles.
- Use `replicate.predictions.get` to poll manually.
