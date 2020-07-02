#!/usr/bin/env ruby
#
# An example of funneling metrics to Redis pub/sub.
#
# to run this, you may need to:
#   gem install redis
#
require 'redis'

r = Redis.new(host: "127.0.0.1", port: 6379, db: 1)

loop do
  # example input: "counter_ruby count=0 1591741648101185000"
  line = STDIN.readline.chomp

  key = line.split(" ")[0]
  key = key.split(",")[0]
  r.publish(key, line)
end
