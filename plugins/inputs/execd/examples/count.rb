#!/usr/bin/env ruby

## Example in Ruby not using any signaling

counter = 0

def time_ns_str(t)
  ns = t.nsec.to_s
  (9 - ns.size).times do 
    ns = "0" + ns # left pad
  end
  t.to_i.to_s + ns
end

loop do
  puts "counter_ruby count=#{counter} #{time_ns_str(Time.now)}"
  STDOUT.flush
  counter += 1

  sleep 1
end
