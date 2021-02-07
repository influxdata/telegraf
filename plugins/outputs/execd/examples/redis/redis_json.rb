#!/usr/bin/env ruby
#
# An example of funneling metrics to Redis pub/sub.
#
# to run this, you may need to:
#   gem install redis
#
require 'redis'
require 'json'

r = Redis.new(host: "127.0.0.1", port: 6379, db: 1)

loop do
  # example input: "{"fields":{"count":0},"name":"counter_ruby","tags":{"host":"localhost"},"timestamp":1586374982}"
  line = STDIN.readline.chomp

  l = JSON.parse(line)

  key = l["name"]
  r.publish(key, line)
end
