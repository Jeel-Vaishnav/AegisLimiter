package redis

const (
	TokenBucketLua = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "last_updated")
local tokens = tonumber(data[1])
local last_updated = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    last_updated = now
else
    local elapsed = math.max(0, now - last_updated)
    local generated_tokens = elapsed * refill_rate
    tokens = math.min(capacity, tokens + generated_tokens)
    last_updated = now
end

local allowed = 0
local retry_after_ms = 0

if tokens >= requested then
    allowed = 1
    tokens = tokens - requested
    retry_after_ms = 0
else
    allowed = 0
    local needed = requested - tokens
    if refill_rate > 0 then
        retry_after_ms = math.ceil(needed / refill_rate)
    else
        retry_after_ms = 3600000
    end
end

redis.call("HMSET", key, "tokens", tokens, "last_updated", last_updated)

local time_to_fill_ms = 1000
if refill_rate > 0 then
    time_to_fill_ms = math.ceil((capacity - tokens) / refill_rate)
end
local ttl_seconds = math.max(60, math.ceil((time_to_fill_ms + 10000) / 1000))
redis.call("EXPIRE", key, ttl_seconds)

return { allowed, math.floor(tokens), retry_after_ms, capacity, time_to_fill_ms }
`

	SlidingWindowCounterLua = `
local base_key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_size_ms = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local current_bucket = math.floor(now / window_size_ms)
local current_bucket_key = base_key .. ":" .. current_bucket
local prev_bucket_key = base_key .. ":" .. (current_bucket - 1)

local current_count = tonumber(redis.call("GET", current_bucket_key) or "0")
local prev_count = tonumber(redis.call("GET", prev_bucket_key) or "0")

local elapsed_in_window = now % window_size_ms
local weight_prev = (window_size_ms - elapsed_in_window) / window_size_ms
local estimated_requests = math.floor(prev_count * weight_prev) + current_count

local allowed = 0
local remaining = 0
local retry_after_ms = 0

if estimated_requests + requested <= limit then
    allowed = 1
    current_count = redis.call("INCRBY", current_bucket_key, requested)
    local ttl_sec = math.ceil((window_size_ms * 2) / 1000)
    redis.call("EXPIRE", current_bucket_key, ttl_sec)
    remaining = math.max(0, limit - (estimated_requests + requested))
    retry_after_ms = 0
else
    allowed = 0
    remaining = 0
    local reset_remaining = window_size_ms - elapsed_in_window
    retry_after_ms = math.max(1, reset_remaining)
end

local reset_ms = window_size_ms - elapsed_in_window
return { allowed, remaining, retry_after_ms, limit, reset_ms }
`

	SlidingWindowLogLua = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_size_ms = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local nonce = ARGV[5] or tostring(now)

local clear_before = now - window_size_ms
redis.call("ZREMRANGEBYSCORE", key, "-inf", clear_before)

local current_count = redis.call("ZCARD", key)
local allowed = 0
local remaining = 0
local retry_after_ms = 0

if current_count + requested <= limit then
    allowed = 1
    for i = 1, requested do
        local member = nonce .. ":" .. i
        redis.call("ZADD", key, now, member)
    end
    remaining = limit - (current_count + requested)
    retry_after_ms = 0
else
    allowed = 0
    remaining = 0
    local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
    if #oldest >= 2 then
        local oldest_time = tonumber(oldest[2])
        retry_after_ms = math.max(1, (oldest_time + window_size_ms) - now)
    else
        retry_after_ms = window_size_ms
    end
end

local ttl_seconds = math.ceil(window_size_ms / 1000) + 5
redis.call("EXPIRE", key, ttl_seconds)

return { allowed, remaining, retry_after_ms, limit, window_size_ms }
`

	LeakyBucketLua = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local leak_rate = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "water", "last_leak")
local water = tonumber(data[1])
local last_leak = tonumber(data[2])

if water == nil then
    water = 0
    last_leak = now
else
    local elapsed = math.max(0, now - last_leak)
    local leaked = elapsed * leak_rate
    water = math.max(0, water - leaked)
    last_leak = now
end

local allowed = 0
local remaining = 0
local retry_after_ms = 0

if water + requested <= capacity then
    allowed = 1
    water = water + requested
    remaining = math.floor(capacity - water)
    retry_after_ms = 0
else
    allowed = 0
    remaining = 0
    local overflow = (water + requested) - capacity
    if leak_rate > 0 then
        retry_after_ms = math.ceil(overflow / leak_rate)
    else
        retry_after_ms = 60000
    end
end

redis.call("HMSET", key, "water", water, "last_leak", last_leak)

local drain_time_ms = 1000
if leak_rate > 0 then
    drain_time_ms = math.ceil(water / leak_rate)
end
local ttl_seconds = math.max(60, math.ceil((drain_time_ms + 10000) / 1000))
redis.call("EXPIRE", key, ttl_seconds)

return { allowed, remaining, retry_after_ms, capacity, drain_time_ms }
`
)
