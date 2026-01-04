function apiBase() {
  return env("OPENROUTER_BASE_URL") || "https://openrouter.ai/api/v1";
}

function authHeaders() {
  const headers = { Authorization: "Bearer " + secret("token") };
  const referer = env("OPENROUTER_REFERER");
  const title = env("OPENROUTER_TITLE");
  if (referer) headers["HTTP-Referer"] = referer;
  if (title) headers["X-Title"] = title;
  return headers;
}

function parseJSON(value, fallback) {
  if (!value) return fallback;
  return JSON.parse(value);
}

function buildMessages(params) {
  if (params.messages) return parseJSON(params.messages, []);
  const out = [];
  if (params.system) {
    out.push({ role: "system", content: params.system });
  }
  if (params.q) {
    out.push({ role: "user", content: params.q });
  }
  return out;
}

export default {
  chat: {
    desc: "Chat completion",
    args: ["q", "model", "system", "messages", "temperature", "top_p", "max_tokens", "stream"],
    run: (params) => {
      const body = {
        model: params.model || env("OPENROUTER_MODEL") || "openai/gpt-4o-mini",
        messages: buildMessages(params),
        temperature: params.temperature ? Number(params.temperature) : undefined,
        top_p: params.top_p ? Number(params.top_p) : undefined,
        max_tokens: params.max_tokens ? Number(params.max_tokens) : undefined,
        stream: params.stream === "true",
      };
      return fetch(apiBase() + "/chat/completions", {
        method: "POST",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
  "models.list": {
    desc: "List models",
    args: [],
    run: () => fetch(apiBase() + "/models", { headers: authHeaders() }),
  },
};
