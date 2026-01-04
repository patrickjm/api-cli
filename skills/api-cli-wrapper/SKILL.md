---
name: api-cli-wrapper
description: Use when creating a new REST API CLI wrapper using this repo's `api` tool, including authoring provider scripts, defining commands, setting secrets/env, and testing with the `api` CLI.
---

# Create a REST API CLI wrapper with `api`

## Workflow

0) Install the CLI via Homebrew:

```bash
brew install patrickmoriarty/tap/api
```

1) Add a provider script under `providers/NAME.js`.
2) Use the required format:

```js
export default {
  "resource.action": {
    desc: "Short description",
    args: [
      { name: "param", desc: "What it does", required: false },
    ],
    run: async (params) => {
      // Use fetch(url, { method, headers, body })
      // Params are strings from --param key=value.
      return fetch("https://api.example.com/v1/resource", {
        method: "GET",
        headers: { Authorization: `Bearer ${secret("token")}` },
      });
    },
  },
};
```

3) Install the provider:

```bash
api install ./providers/NAME.js --name NAME
```

4) Set secrets/env per profile:

```bash
api secret set NAME token "<API_KEY>"
api env set NAME BASE_URL "https://api.example.com"
```

5) Inspect commands:

```bash
api inspect NAME --json
```

6) Run commands:

```bash
api NAME.resource.action --param foo=bar --json
```

## Helpers available in provider scripts

- `secret(name)` returns a secret for the active profile.
- `env(name, fallback)` returns profile env value or OS env.
- `fetch(url, opts)` or `fetch(opts)` for HTTP requests.
- `sleep(ms)` for polling loops.

## Patterns

- Centralize base URLs: `const base = env("BASE_URL", "https://api.example.com");`
- Use JSON requests: `body: JSON.stringify(payload)` and `headers: { "content-type": "application/json" }`.
- For long jobs, add `*.wait` that polls with `sleep`.
