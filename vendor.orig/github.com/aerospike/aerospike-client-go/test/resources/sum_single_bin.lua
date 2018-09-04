local function reducer(val1,val2)
	return val1 + val2
end

function sum_single_bin(stream,name)
	local function mapper(rec)
		return rec[name]
	end

	return stream : map(mapper) : reduce(reducer)
end
