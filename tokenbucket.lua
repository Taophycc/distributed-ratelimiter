local key      = KEYS[1]
local capacity = tonumber(ARGV[1])
local interval = tonumber(ARGV[2])  -- ms per token


if not capacity or capacity < 1 or not interval or interval < 1 then
    return redis.error_reply("capacity and interval must be positive")
end

local t   = redis.call("TIME")
local now = t[1] * 1000 + math.floor(t[2] / 1000)

local data   = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(data[1])
local last   = tonumber(data[2])
if tokens == nil then
    tokens = capacity
    last   = now
end

local earned = math.floor((now - last) / interval)
if earned > 0 then
    tokens = math.min(capacity, tokens + earned)
    last   = last + earned * interval
end

local allowed = 0
if tokens > 0 then
    tokens  = tokens - 1
    allowed = 1
end

redis.call("HSET", key, "tokens", tokens, "last_refill", last)
redis.call("PEXPIRE", key, capacity * interval)

local reset_at = now
if tokens == 0 then
    reset_at = last + interval
end
return {allowed, tokens, reset_at}