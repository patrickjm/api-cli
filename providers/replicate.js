function apiBase() {
  return env("REPLICATE_BASE_URL") || "https://api.replicate.com/v1";
}

function authHeaders() {
  return { Authorization: "Bearer " + secret("token") };
}

function parseJSON(value, fallback) {
  if (!value) return fallback;
  return JSON.parse(value);
}

function qs(params) {
  const parts = [];
  for (const key in params) {
    if (params[key] === undefined || params[key] === null || params[key] === "") continue;
    parts.push(encodeURIComponent(key) + "=" + encodeURIComponent(params[key]));
  }
  return parts.length ? "?" + parts.join("&") : "";
}

function waitHeader(params) {
  if (!params.wait) return null;
  if (params.wait === "true") return "wait=60";
  return "wait=" + params.wait;
}

function fetchJSON(path, opts) {
  const resp = fetch(apiBase() + path, opts);
  return resp.json;
}

export default {
  search: {
    desc: "Search models, collections, docs",
    args: ["q", "limit"],
    run: (params) => fetchJSON("/search" + qs({ query: params.q, limit: params.limit }), { headers: authHeaders() }),
  },
  "models.list": {
    desc: "List models",
    args: ["cursor", "sort_by", "sort_direction"],
    run: (params) => fetchJSON("/models" + qs({
      cursor: params.cursor,
      sort_by: params.sort_by,
      sort_direction: params.sort_direction,
    }), { headers: authHeaders() }),
  },
  "models.get": {
    desc: "Get model",
    args: ["owner", "name"],
    run: (params) => fetchJSON("/models/" + params.owner + "/" + params.name, { headers: authHeaders() }),
  },
  "models.examples": {
    desc: "List model examples",
    args: ["owner", "name"],
    run: (params) => fetchJSON("/models/" + params.owner + "/" + params.name + "/examples", { headers: authHeaders() }),
  },
  "models.versions": {
    desc: "List model versions",
    args: ["owner", "name"],
    run: (params) => fetchJSON("/models/" + params.owner + "/" + params.name + "/versions", { headers: authHeaders() }),
  },
  "models.version": {
    desc: "Get model version",
    args: ["owner", "name", "version"],
    run: (params) => fetchJSON("/models/" + params.owner + "/" + params.name + "/versions/" + params.version, { headers: authHeaders() }),
  },
  "predictions.create": {
    desc: "Create prediction",
    args: ["version", "input", "wait", "cancel_after", "webhook", "webhook_events_filter"],
    run: (params) => {
      const headers = Object.assign({ "Content-Type": "application/json" }, authHeaders());
      const prefer = waitHeader(params);
      if (prefer) headers.Prefer = prefer;
      if (params.cancel_after) headers["Cancel-After"] = params.cancel_after;
      return fetchJSON("/predictions", {
        method: "POST",
        headers: headers,
        body: {
          version: params.version,
          input: parseJSON(params.input, undefined),
          webhook: params.webhook,
          webhook_events_filter: parseJSON(params.webhook_events_filter, undefined),
        },
      });
    },
  },
  "predictions.get": {
    desc: "Get prediction",
    args: ["id"],
    run: (params) => fetchJSON("/predictions/" + params.id, { headers: authHeaders() }),
  },
  "predictions.cancel": {
    desc: "Cancel prediction",
    args: ["id"],
    run: (params) => fetchJSON("/predictions/" + params.id + "/cancel", {
      method: "POST",
      headers: authHeaders(),
    }),
  },
  "predictions.wait": {
    desc: "Poll prediction until done",
    args: ["id", "poll_ms", "timeout_s"],
    run: (params) => {
      const pollMs = params.poll_ms ? Number(params.poll_ms) : 2000;
      const timeoutMs = params.timeout_s ? Number(params.timeout_s) * 1000 : 300000;
      const start = Date.now();
      while (true) {
        const pred = fetchJSON("/predictions/" + params.id, { headers: authHeaders() });
        if (pred.status === "succeeded" || pred.status === "failed" || pred.status === "canceled") {
          return pred;
        }
        if (Date.now() - start > timeoutMs) {
          return pred;
        }
        sleep(pollMs);
      }
    },
  },
};
