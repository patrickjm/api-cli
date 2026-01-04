function apiBase() {
  return env("PERPLEXITY_BASE_URL") || "https://api.perplexity.ai";
}

function authHeaders() {
  return { Authorization: "Bearer " + secret("token") };
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
  search: {
    desc: "Perplexity search",
    args: ["q", "max_results", "max_tokens", "max_tokens_per_page", "search_domain_filter", "country", "search_recency_filter", "search_after_date", "search_before_date"],
    run: (params) => {
      const body = {
        query: params.q || parseJSON(params.query, ""),
        max_results: params.max_results ? Number(params.max_results) : undefined,
        max_tokens: params.max_tokens ? Number(params.max_tokens) : undefined,
        max_tokens_per_page: params.max_tokens_per_page ? Number(params.max_tokens_per_page) : undefined,
        search_domain_filter: parseJSON(params.search_domain_filter, undefined),
        country: params.country,
        search_recency_filter: params.search_recency_filter,
        search_after_date: params.search_after_date,
        search_before_date: params.search_before_date,
      };
      return fetch(apiBase() + "/search", {
        method: "POST",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
  ask: {
    desc: "Chat completion with sonar",
    args: ["q", "model", "system", "messages", "search_mode", "temperature", "top_p", "max_tokens", "return_images", "return_related_questions", "search_domain_filter", "search_recency_filter", "search_after_date_filter", "search_before_date_filter", "last_updated_after_filter", "last_updated_before_filter"],
    run: (params) => {
      const body = {
        model: params.model || "sonar-pro",
        messages: buildMessages(params),
        search_mode: params.search_mode || "web",
        temperature: params.temperature ? Number(params.temperature) : 0.2,
        top_p: params.top_p ? Number(params.top_p) : 0.9,
        max_tokens: params.max_tokens ? Number(params.max_tokens) : undefined,
        return_images: params.return_images === "true",
        return_related_questions: params.return_related_questions === "true",
        search_domain_filter: parseJSON(params.search_domain_filter, undefined),
        search_recency_filter: params.search_recency_filter,
        search_after_date_filter: params.search_after_date_filter,
        search_before_date_filter: params.search_before_date_filter,
        last_updated_after_filter: params.last_updated_after_filter,
        last_updated_before_filter: params.last_updated_before_filter,
      };
      return fetch(apiBase() + "/chat/completions", {
        method: "POST",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
  deep: {
    desc: "Deep research chat completion",
    args: ["q", "model", "system", "messages", "search_mode", "reasoning_effort", "temperature", "top_p", "max_tokens", "return_images", "return_related_questions", "search_domain_filter", "search_recency_filter", "search_after_date_filter", "search_before_date_filter", "last_updated_after_filter", "last_updated_before_filter"],
    run: (params) => {
      const body = {
        model: params.model || "sonar-deep-research",
        messages: buildMessages(params),
        search_mode: params.search_mode || "web",
        reasoning_effort: params.reasoning_effort || "medium",
        temperature: params.temperature ? Number(params.temperature) : 0.2,
        top_p: params.top_p ? Number(params.top_p) : 0.9,
        max_tokens: params.max_tokens ? Number(params.max_tokens) : undefined,
        return_images: params.return_images === "true",
        return_related_questions: params.return_related_questions === "true",
        search_domain_filter: parseJSON(params.search_domain_filter, undefined),
        search_recency_filter: params.search_recency_filter,
        search_after_date_filter: params.search_after_date_filter,
        search_before_date_filter: params.search_before_date_filter,
        last_updated_after_filter: params.last_updated_after_filter,
        last_updated_before_filter: params.last_updated_before_filter,
      };
      return fetch(apiBase() + "/chat/completions", {
        method: "POST",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
};
