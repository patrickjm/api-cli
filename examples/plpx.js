export default {
  default: {
    desc: "Perplexity chat completion",
    args: ["q"],
    run: (params) => {
      return fetch("https://api.perplexity.ai/chat/completions", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${secret("token")}`,
          "Content-Type": "application/json",
        },
        body: {
          model: "sonar-pro",
          messages: [
            {
              role: "user",
              content: params.q,
            },
          ],
        },
      });
    },
  },
};
