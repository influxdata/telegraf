
assert(rawget(_G, "stat") == nil)  -- module not loaded before

if T == nil then
  stat = function () print"`querytab' nao ativo" end
  return
end


function checktable (t)
  local asize, hsize, ff = T.querytab(t)
  local l = {}
  for i=0,hsize-1 do
    local key,val,next = T.querytab(t, i + asize)
    if key == nil then
      assert(l[i] == nil and val==nil and next==nil)
    elseif key == "<undef>" then
      assert(val==nil)
    else
      assert(t[key] == val)
      local mp = T.hash(key, t)
      if l[i] then
        assert(l[i] == mp)
      elseif mp ~= i then
        l[i] = mp
      else  -- list head
        l[mp] = {mp}   -- first element
        while next do
          assert(ff <= next and next < hsize)
          if l[next] then assert(l[next] == mp) else l[next] = mp end
          table.insert(l[mp], next)
          key,val,next = T.querytab(t, next)
          assert(key)
        end
      end
    end
  end
  l.asize = asize; l.hsize = hsize; l.ff = ff
  return l
end

function mostra (t)
  local asize, hsize, ff = T.querytab(t)
  print(asize, hsize, ff)
  print'------'
  for i=0,asize-1 do
    local _, v = T.querytab(t, i)
    print(string.format("[%d] -", i), v)
  end
  print'------'
  for i=0,hsize-1 do
    print(i, T.querytab(t, i+asize))
  end
  print'-------------'
end

function stat (t)
  t = checktable(t)
  local nelem, nlist = 0, 0
  local maxlist = {}
  for i=0,t.hsize-1 do
    if type(t[i]) == 'table' then
      local n = table.getn(t[i])
      nlist = nlist+1
      nelem = nelem + n
      if not maxlist[n] then maxlist[n] = 0 end
      maxlist[n] = maxlist[n]+1
    end
  end
  print(string.format("hsize=%d  elements=%d  load=%.2f  med.len=%.2f (asize=%d)",
          t.hsize, nelem, nelem/t.hsize, nelem/nlist, t.asize))
  for i=1,table.getn(maxlist) do
    local n = maxlist[i] or 0
    print(string.format("%5d %10d %.2f%%", i, n, n*100/nlist))
  end
end

