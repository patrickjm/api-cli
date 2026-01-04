Provider Scripts

Install
- api install ./providers/alpaca.js --name alpaca
- api install ./providers/perplexity.js --name perplexity
- api install ./providers/replicate.js --name replicate
- api install ./providers/openrouter.js --name openrouter

Alpaca
- Secrets: key, secret
- Env: ALPACA_BASE_URL (default paper), ALPACA_DATA_BASE_URL (default data)
- Example: api alpaca.orders.list -s status=open
- Example: api alpaca.orders.create -s symbol=AAPL -s qty=1 -s side=buy

Perplexity
- Secret: token
- Env: PERPLEXITY_BASE_URL
- Example: api perplexity.search -s q="latest AI developments"
- Example: api perplexity.ask -s q="Stock market overview" -s model=sonar-pro
- Example: api perplexity.deep -s q="Deep research on lithium supply chain" -s reasoning_effort=high

Replicate
- Secret: token
- Env: REPLICATE_BASE_URL
- Example: api replicate.search -s q="sdxl"
- Example: api replicate.predictions.create -s version=replicate/hello-world:5c7d... -s input='{"text":"Alice"}' --json
- Example: api replicate.predictions.wait -s id=<prediction_id> -s poll_ms=2000 --json

OpenRouter
- Secret: token
- Env: OPENROUTER_MODEL, OPENROUTER_REFERER, OPENROUTER_TITLE, OPENROUTER_BASE_URL
- Example: api openrouter.chat -s q="What is the meaning of life?" -s model=openai/gpt-4o-mini
