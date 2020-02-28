#!/usr/bin/env ruby

## Example in Ruby not using any signaling

counter = 0

loop do
  puts "counter_ruby count=#{counter}"
  STDOUT.flush
  counter += 1

  sleep 1
end
