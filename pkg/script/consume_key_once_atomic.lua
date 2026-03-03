local exists = redis.call("EXISTS", KEYS[1])
if exists == 1 then
    redis.call("DEL", KEYS[1])
    return 1
end
return 0
