-- Leaky Bucket Rate Limiter Lua Script
-- Models a bucket with water level and constant leak rate
--
-- Keys:
--   KEYS[1]: rate limiter key (e.g., "ratelimit:leaky:user_123")
-- Arguments:
--   ARGV[1]: capacity (bucket depth / maximum queued requests)
--   ARGV[2]: leak_rate (leaked water/requests per millisecond)
--   ARGV[3]: requested (water units added, usually 1)
--   ARGV[4]: current timestamp in milliseconds

local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local leak_rate = tonumber(ARGV[2]) -- requests per millisecond
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
