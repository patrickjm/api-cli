function baseUrl() {
  return env("ALPACA_BASE_URL") || env("ALPACA_ENDPOINT") || "https://paper-api.alpaca.markets";
}

function dataBaseUrl() {
  return env("ALPACA_DATA_BASE_URL") || "https://data.alpaca.markets";
}

function authHeaders() {
  return {
    "APCA-API-KEY-ID": secret("key"),
    "APCA-API-SECRET-KEY": secret("secret"),
  };
}

function qs(params) {
  const parts = [];
  for (const key in params) {
    if (params[key] === undefined || params[key] === null || params[key] === "") continue;
    parts.push(encodeURIComponent(key) + "=" + encodeURIComponent(params[key]));
  }
  return parts.length ? "?" + parts.join("&") : "";
}

export default {
  "account.get": {
    desc: "Get account details",
    args: [],
    run: () => fetch(baseUrl() + "/v2/account", { headers: authHeaders() }),
  },
  "assets.list": {
    desc: "List assets",
    args: ["status", "asset_class", "exchange"],
    run: (params) => fetch(baseUrl() + "/v2/assets" + qs({
      status: params.status,
      asset_class: params.asset_class,
      exchange: params.exchange,
    }), { headers: authHeaders() }),
  },
  "assets.get": {
    desc: "Get asset by id or symbol",
    args: ["id"],
    run: (params) => fetch(baseUrl() + "/v2/assets/" + params.id, { headers: authHeaders() }),
  },
  "clock": {
    desc: "Get market clock",
    args: [],
    run: () => fetch(baseUrl() + "/v2/clock", { headers: authHeaders() }),
  },
  "calendar": {
    desc: "Get market calendar",
    args: ["start", "end"],
    run: (params) => fetch(baseUrl() + "/v2/calendar" + qs({ start: params.start, end: params.end }), { headers: authHeaders() }),
  },
  "orders.list": {
    desc: "List orders",
    args: ["status", "limit", "after", "until", "direction", "nested"],
    run: (params) => fetch(baseUrl() + "/v2/orders" + qs({
      status: params.status,
      limit: params.limit,
      after: params.after,
      until: params.until,
      direction: params.direction,
      nested: params.nested,
    }), { headers: authHeaders() }),
  },
  "orders.get": {
    desc: "Get an order",
    args: ["id"],
    run: (params) => fetch(baseUrl() + "/v2/orders/" + params.id, { headers: authHeaders() }),
  },
  "orders.create": {
    desc: "Create an order",
    args: ["symbol", "qty", "notional", "side", "type", "time_in_force", "limit_price", "stop_price", "trail_price", "trail_percent", "extended_hours", "client_order_id", "order_class", "take_profit", "stop_loss"],
    run: (params) => {
      const body = {
        symbol: params.symbol,
        qty: params.qty,
        notional: params.notional,
        side: params.side || "buy",
        type: params.type || "market",
        time_in_force: params.time_in_force || "day",
        limit_price: params.limit_price,
        stop_price: params.stop_price,
        trail_price: params.trail_price,
        trail_percent: params.trail_percent,
        extended_hours: params.extended_hours === "true",
        client_order_id: params.client_order_id,
        order_class: params.order_class,
        take_profit: params.take_profit ? JSON.parse(params.take_profit) : undefined,
        stop_loss: params.stop_loss ? JSON.parse(params.stop_loss) : undefined,
      };
      return fetch(baseUrl() + "/v2/orders", {
        method: "POST",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
  "orders.replace": {
    desc: "Replace an order",
    args: ["id", "qty", "time_in_force", "limit_price", "stop_price", "trail", "client_order_id"],
    run: (params) => {
      const body = {
        qty: params.qty,
        time_in_force: params.time_in_force,
        limit_price: params.limit_price,
        stop_price: params.stop_price,
        trail: params.trail,
        client_order_id: params.client_order_id,
      };
      return fetch(baseUrl() + "/v2/orders/" + params.id, {
        method: "PATCH",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
  "orders.cancel": {
    desc: "Cancel an order",
    args: ["id"],
    run: (params) => fetch(baseUrl() + "/v2/orders/" + params.id, {
      method: "DELETE",
      headers: authHeaders(),
    }),
  },
  "positions.list": {
    desc: "List positions",
    args: [],
    run: () => fetch(baseUrl() + "/v2/positions", { headers: authHeaders() }),
  },
  "positions.get": {
    desc: "Get a position",
    args: ["symbol"],
    run: (params) => fetch(baseUrl() + "/v2/positions/" + params.symbol, { headers: authHeaders() }),
  },
  "positions.close": {
    desc: "Close a position",
    args: ["symbol"],
    run: (params) => fetch(baseUrl() + "/v2/positions/" + params.symbol, {
      method: "DELETE",
      headers: authHeaders(),
    }),
  },
  "activities.list": {
    desc: "List account activities",
    args: ["activity_types", "date", "until", "after", "direction", "page_size", "page_token"],
    run: (params) => fetch(baseUrl() + "/v2/account/activities" + qs({
      activity_types: params.activity_types,
      date: params.date,
      until: params.until,
      after: params.after,
      direction: params.direction,
      page_size: params.page_size,
      page_token: params.page_token,
    }), { headers: authHeaders() }),
  },
  "watchlists.list": {
    desc: "List watchlists",
    args: [],
    run: () => fetch(baseUrl() + "/v2/watchlists", { headers: authHeaders() }),
  },
  "watchlists.get": {
    desc: "Get a watchlist",
    args: ["id"],
    run: (params) => fetch(baseUrl() + "/v2/watchlists/" + params.id, { headers: authHeaders() }),
  },
  "watchlists.create": {
    desc: "Create a watchlist",
    args: ["name", "symbols"],
    run: (params) => {
      const body = {
        name: params.name,
        symbols: params.symbols ? params.symbols.split(",") : undefined,
      };
      return fetch(baseUrl() + "/v2/watchlists", {
        method: "POST",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
  "watchlists.add": {
    desc: "Add a symbol to a watchlist",
    args: ["id", "symbol"],
    run: (params) => {
      const body = { symbol: params.symbol };
      return fetch(baseUrl() + "/v2/watchlists/" + params.id, {
        method: "POST",
        headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
        body: body,
      });
    },
  },
  "watchlists.delete": {
    desc: "Delete a watchlist",
    args: ["id"],
    run: (params) => fetch(baseUrl() + "/v2/watchlists/" + params.id, {
      method: "DELETE",
      headers: authHeaders(),
    }),
  },
  "data.stocks.quote": {
    desc: "Get latest stock quote",
    args: ["symbol"],
    run: (params) => fetch(dataBaseUrl() + "/v2/stocks/" + params.symbol + "/quotes/latest", { headers: authHeaders() }),
  },
  "data.stocks.trade": {
    desc: "Get latest stock trade",
    args: ["symbol"],
    run: (params) => fetch(dataBaseUrl() + "/v2/stocks/" + params.symbol + "/trades/latest", { headers: authHeaders() }),
  },
  "data.stocks.bars": {
    desc: "Get stock bars",
    args: ["symbol", "timeframe", "start", "end", "limit", "adjustment"],
    run: (params) => fetch(dataBaseUrl() + "/v2/stocks/" + params.symbol + "/bars" + qs({
      timeframe: params.timeframe || "1Day",
      start: params.start,
      end: params.end,
      limit: params.limit,
      adjustment: params.adjustment,
    }), { headers: authHeaders() }),
  },
};
