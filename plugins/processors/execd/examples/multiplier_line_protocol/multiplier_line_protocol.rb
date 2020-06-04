#!/usr/bin/env ruby

loop do
  # example input: "counter_ruby count=0 1586302128978187000"
  line = STDIN.readline.chomp
  # parse out influx line protocol sections with a really simple hand-rolled parser that doesn't support escaping.
  # for a full line parser in ruby, check out something like the influxdb-lineprotocol-parser gem.
  parts = line.split(" ")
  case parts.size
  when 3
    measurement, fields, timestamp = parts
  when 4
    measurement, tags, fields, timestamp = parts
  else
    STDERR.puts "Unable to parse line protocol"
    exit 1
  end
  fields = fields.split(",").map{|t| 
    k,v = t.split("=") 
    if k == "count"
      v = v.to_i * 2 # multiple count metric by two
    end
    "#{k}=#{v}"
  }.join(",")
  puts [measurement, tags, fields, timestamp].select{|s| s && s.size != 0 }.join(" ")
  STDOUT.flush
end
