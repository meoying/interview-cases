local key = KEYS[1]
local limit = tonumber(ARGV[1])
-- 假设 window 是毫秒表达的
local window = tonumber(ARGV[2])
-- 当前时间戳我们要求服务端传过来，因为本质上这个窗口描述的是服务端，
-- 只是借助了 Redis 来实现，同样是毫秒表达
-- 这种做法有一个弊端，就是可能先到服务端到 Redis 之间网络传输时间难确定
-- 但是一般我们限流不在意这么一点不精确
local now = tonumber(ARGV[3])
-- 窗口的最小时间，时间小于这个，就说明已经不在窗口内了
local windowStart = now - window


-- 获得 list 的长度
local len = tonumber(redis.call('LLEN', key))
-- 这是一个小优化，避免每个请求都触发淘汰
if(len >= limit) then
    -- list 满了，现在要淘汰了
    local head = tonumber(redis.call('LPOP', key))
    while head <= windowStart do
        head = tonumber(redis.call('LPOP', key))
    end
    -- 说明当前的 head 还在窗口内，放回去
    redis.call('LPUSH', key, head)
    -- 看看淘汰不在窗口内请求之后，还有多少个
    -- 实际上你可以在 while 循环里面手动维护 len
    -- 我这里为了可读性就再次调用了 LLEN
    len = tonumber(redis.call('LLEN', key))
end

if(len < limit) then
    -- 插入这个请求的时间戳
    redis.call('RPUSH', key, now)
    redis.call('PEXPIRE', key, window)
    -- 允许
    return 1
end
-- 说明超过阈值了
return 0