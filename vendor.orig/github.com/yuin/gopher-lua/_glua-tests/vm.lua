for i, v in ipairs({"hoge", {}, function() end, true, nil}) do
  local ok, msg = pcall(function()
    print(-v)
  end)
  assert(not ok and string.find(msg, "__unm undefined"))
end

assert(#"abc" == 3)
local tbl = {1,2,3}
setmetatable(tbl, {__len = function(self)
  return 10
end})
assert(#tbl == 10)

setmetatable(tbl, nil)
assert(#tbl == 3)

local ok, msg = pcall(function()
  return 1 < "hoge"
end)
assert(not ok and string.find(msg, "attempt to compare number with string"))

local ok, msg = pcall(function()
  return {} < (function() end)
end)
assert(not ok and string.find(msg, "attempt to compare table with function"))

local ok, msg = pcall(function()
  for n = nil,1 do
    print(1)
  end
end)
assert(not ok and string.find(msg, "for statement init must be a number"))

local ok, msg = pcall(function()
  for n = 1,nil do
    print(1)
  end
end)
assert(not ok and string.find(msg, "for statement limit must be a number"))

local ok, msg = pcall(function()
  for n = 1,10,nil do
    print(1)
  end
end)
assert(not ok and string.find(msg, "for statement step must be a number"))

local ok, msg = pcall(function()
  return {} + (function() end)
end)
assert(not ok and string.find(msg, "cannot perform add operation between table and function"))

local ok, msg = pcall(function()
  return {} .. (function() end)
end)
assert(not ok and string.find(msg, "cannot perform concat operation between table and function"))

-- test table with initial elements over 511
local bigtable = {1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
                  1,1,1,1,1,1,1,1,1,1,1,1,10}
assert(bigtable[601] == 10)

local ok, msg = loadstring([[
  function main(
     a1,  a2,  a3,  a4,  a5,  a6,  a7,  a8,
     a9,  a10,  a11,  a12,  a13,  a14,  a15,  a16,
     a17,  a18,  a19,  a20,  a21,  a22,  a23,  a24,
     a25,  a26,  a27,  a28,  a29,  a30,  a31,  a32,
     a33,  a34,  a35,  a36,  a37,  a38,  a39,  a40,
     a41,  a42,  a43,  a44,  a45,  a46,  a47,  a48,
     a49,  a50,  a51,  a52,  a53,  a54,  a55,  a56,
     a57,  a58,  a59,  a60,  a61,  a62,  a63,  a64,
     a65,  a66,  a67,  a68,  a69,  a70,  a71,  a72,
     a73,  a74,  a75,  a76,  a77,  a78,  a79,  a80,
     a81,  a82,  a83,  a84,  a85,  a86,  a87,  a88,
     a89,  a90,  a91,  a92,  a93,  a94,  a95,  a96,
     a97,  a98,  a99,  a100,  a101,  a102,  a103,
     a104,  a105,  a106,  a107,  a108,  a109,  a110,
     a111,  a112,  a113,  a114,  a115,  a116,  a117,
     a118,  a119,  a120,  a121,  a122,  a123,  a124,
     a125,  a126,  a127,  a128,  a129,  a130,  a131,
     a132,  a133,  a134,  a135,  a136,  a137,  a138,
     a139,  a140,  a141,  a142,  a143,  a144,  a145,
     a146,  a147,  a148,  a149,  a150,  a151,  a152,
     a153,  a154,  a155,  a156,  a157,  a158,  a159,
     a160,  a161,  a162,  a163,  a164,  a165,  a166,
     a167,  a168,  a169,  a170,  a171,  a172,  a173,
     a174,  a175,  a176,  a177,  a178,  a179,  a180,
     a181,  a182,  a183,  a184,  a185,  a186,  a187,
     a188,  a189,  a190,  a191,  a192,  a193,  a194,
     a195,  a196,  a197,  a198,  a199,  a200,  a201,
     a202,  a203,  a204,  a205,  a206,  a207,  a208,
     a209,  a210,  a211,  a212,  a213,  a214,  a215,
     a216,  a217,  a218,  a219,  a220,  a221,  a222,
     a223,  a224,  a225,  a226,  a227,  a228,  a229,
     a230,  a231,  a232,  a233,  a234,  a235,  a236,
     a237,  a238,  a239,  a240,  a241,  a242,  a243,
     a244,  a245,  a246,  a247,  a248,  a249,  a250,
     a251,  a252,  a253,  a254,  a255,  a256,  a257,
     a258,  a259,  a260,  a261,  a262,  a263,  a264,
     a265,  a266,  a267,  a268,  a269,  a270,  a271,
     a272,  a273,  a274,  a275,  a276,  a277,  a278,
     a279,  a280,  a281,  a282,  a283,  a284,  a285,
     a286,  a287,  a288,  a289,  a290,  a291,  a292,
     a293,  a294,  a295,  a296,  a297,  a298,  a299,
     a300,  a301,  a302) end
]]) 
assert(not ok and string.find(msg, "register overflow"))

local ok, msg = loadstring([[
   function main()
     local a = {...}
   end
]])
assert(not ok and string.find(msg, "cannot use '...' outside a vararg function"))
