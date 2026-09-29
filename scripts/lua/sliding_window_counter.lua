-- Sliding Window Counter Rate Limiter Lua Script
-- Uses a hybrid sliding window counter approximation (Cloudflare / Stripe model)
-- O(1) time complexity and O(1) space complexity per key
--
-- Keys:
--   KEYS[1]: base key for current rate limit bucket (e.g., "ratelimit:sw:user_123")
-- Arguments:
--   ARGV[1]: limit (maximum allowed requests within window)
--   ARGV[2]: window_size_ms (duration of window in milliseconds, e.g., 60000 for 1 min)
--   ARGV[3]: requested count (usually 1)
--   ARGV[4]: current timestamp in milliseconds

local base_key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_size_ms = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

-- Determine current and previous window time buckets
local current_bucket = math.floor(now / window_size_ms)
local current_bucket_key = base_key .. ":" .. current_bucket
local prev_bucket_key = base_key .. ":" .. (current_bucket - 1)

-- Fetch counts from current bucket and previous bucket
local current_count = tonumber(redis.call("GET", current_bucket_key) or "0")
local prev_count = tonumber(redis.call("GET", prev_bucket_key) or "0")

-- Calculate weight of previous window based on elapsed time in current window
-- Progress within current window: (now % window_size_ms) / window_size_ms
local elapsed_in_window = now % window_size_ms
local weight_prev = (window_size_ms - elapsed_in_window) / window_size_ms

-- Estimated total requests in the sliding window
local estimated_requests = math.floor(prev_count * weight_prev) + current_count

local allowed = 0
local remaining = 0
local retry_after_ms = 0

if estimated_requests + requested <= limit then
    -- Allowed: increment current bucket
    allowed = 1
    current_count = redis.call("INCRBY", current_bucket_key, requested)
    
    -- Ensure TTL on the current bucket (window_size * 2 in seconds so prev window remains available)
    local ttl_sec = math.ceil((window_size_ms * 2) / 1000)
    redis.call("EXPIRE", current_bucket_key, ttl_sec)
    
    remaining = math.max(0, limit - (estimated_requests + requested))
    retry_after_ms = 0
else
    -- Rate limit exceeded
    allowed = 0
    remaining = 0
    -- Time until the sliding window advances enough to allow the request
    -- At minimum, wait until the current window rolls over or fraction of prev count drops
    local reset_remaining = window_size_ms - elapsed_in_window
    retry_after_ms = math.max(1, reset_remaining)
end

-- Time until current window ends (reset timestamp in ms)
local reset_ms = window_size_ms - elapsed_in_window

-- Return:
-- 1: allowed (1 or 0)
-- 2: remaining requests in current window
-- 3: retry_after in milliseconds (0 if allowed)
-- 4: limit
-- 5: reset_ms (milliseconds until window reset)
return { allowed, remaining, retry_after_ms, limit, reset_ms }
