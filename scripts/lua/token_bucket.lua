-- Token Bucket Rate Limiter Lua Script
-- Keys:
--   KEYS[1]: rate limiter key (e.g., "ratelimit:token_bucket:user_123")
-- Arguments:
--   ARGV[1]: bucket capacity (maximum burst tokens)
--   ARGV[2]: refill rate (tokens per millisecond)
--   ARGV[3]: requested tokens (cost of current request, usually 1)
--   ARGV[4]: current timestamp in milliseconds

local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2]) -- tokens per millisecond
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

-- Retrieve current state from Redis Hash
local data = redis.call("HMGET", key, "tokens", "last_updated")
local tokens = tonumber(data[1])
local last_updated = tonumber(data[2])

if tokens == nil then
    -- First time seeing this key: initialize to full capacity
    tokens = capacity
    last_updated = now
else
    -- Compute token replenishment since last update
    local elapsed = math.max(0, now - last_updated)
    local generated_tokens = elapsed * refill_rate
    tokens = math.min(capacity, tokens + generated_tokens)
    last_updated = now
end

local allowed = 0
local retry_after_ms = 0

if tokens >= requested then
    -- Consume tokens
    allowed = 1
    tokens = tokens - requested
    retry_after_ms = 0
else
    -- Insufficient tokens: calculate wait time until enough tokens are refilled
    allowed = 0
    local needed = requested - tokens
    if refill_rate > 0 then
        retry_after_ms = math.ceil(needed / refill_rate)
    else
        retry_after_ms = 3600000 -- default 1 hour if refill rate is 0
    end
end

-- Save updated state
redis.call("HMSET", key, "tokens", tokens, "last_updated", last_updated)

-- Auto-expire key after it would naturally refill to full capacity plus buffer
-- Prevents memory leakage for stale or inactive keys
local time_to_fill_ms = 1000
if refill_rate > 0 then
    time_to_fill_ms = math.ceil((capacity - tokens) / refill_rate)
end
local ttl_seconds = math.max(60, math.ceil((time_to_fill_ms + 10000) / 1000))
redis.call("EXPIRE", key, ttl_seconds)

-- Return:
-- 1: allowed (1 for yes, 0 for no)
-- 2: remaining tokens (rounded down)
-- 3: retry_after in milliseconds (0 if allowed)
-- 4: capacity
-- 5: time to fully replenish in milliseconds
return { allowed, math.floor(tokens), retry_after_ms, capacity, time_to_fill_ms }
