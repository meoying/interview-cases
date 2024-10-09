-- 删除原有的key
redis.call('DEL', KEYS[1])
for i=1, #ARGV, 2 do
    redis.call('ZADD', KEYS[1], ARGV[i], ARGV[i+1])
end
return 'OK'