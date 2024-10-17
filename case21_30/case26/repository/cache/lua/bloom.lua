if redis.call("BF.EXISTS", KEYS[1], ARGV[1]) == 1 then
    return false
else
    redis.call("BF.ADD", KEYS[1], ARGV[1])
    return true
end
