function average(s, name)

    local function mapper(out, rec)
        out['sum'] = (out['sum'] or 0) + (rec[name] or 0)
        out['count'] = (out['count'] or 0) + 1
        return out
    end

    local function reducer(a, b)
        local out = map() 

        out['sum'] = a['sum'] + b['sum']
        out['count'] = a['count'] + b['count']
        return out
    end

    return s : aggregate(map{sum = 0, count = 0}, mapper) : reduce(reducer)
end
