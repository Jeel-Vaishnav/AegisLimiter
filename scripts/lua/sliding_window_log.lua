-- Sliding Window Log Rate Limiter Lua Script
-- Uses Redis Sorted Sets (ZSET) for exact 100% precision timestamp logging
--
-- Keys:
--   KEYS[1]: rate limiter key (e.g., "ratelimit:swlog:user_123")
-- Arguments:
--   ARGV[1]: limit (maximum allowed requests)
--   ARGV[2]: window_size_ms (duration of window in milliseconds)
--   ARGV[3]: requested count (cost)
--   ARGV[4]: current timestamp in milliseconds
--   ARGV[5]: unique request identifier / nonce

local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_size_ms = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])
local nonce = ARGV[5] or tostring(now)

local clear_before = now - window_size_ms

-- 1. Remove all log elements older than window threshold
redis.call("ZREMRANGEBYSCORE", key, "-inf", clear_before)

-- 2. Count existing requests in the active window
local current_count = redis.call("ZCARD", key)

local allowed = 0
local remaining = 0
local retry_after_ms = 0

if current_count + requested <= limit then
    -- Allowed: insert the new request entry with score = now
    allowed = 1
    for i = 1, requested do
        local member = nonce .. ":" .. i
        redis.call("ZADD", key, now, member)
    end
    remaining = limit - (current_count + requested)
    retry_after_ms = 0
else
    -- Rejected: find oldest element to determine retry_after
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

-- Refresh key TTL
local ttl_seconds = math.ceil(window_size_ms / 1000) + 5
redis.call("EXPIRE", key, ttl_seconds)

return { allowed, remaining, retry_after_ms, limit, window_size_ms }
