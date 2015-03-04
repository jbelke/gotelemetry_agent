#!/usr/bin/env ruby

require 'json'

result = {
	"value" => ARGV[0].to_i
}

print result.to_json