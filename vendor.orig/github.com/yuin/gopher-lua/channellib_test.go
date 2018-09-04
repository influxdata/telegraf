package lua

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestChannelMake(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfScriptFail(t, L, `
    ch = channel.make()
    `)
	obj := L.GetGlobal("ch")
	ch, ok := obj.(LChannel)
	errorIfFalse(t, ok, "channel expected")
	errorIfNotEqual(t, 0, reflect.ValueOf(ch).Cap())
	close(ch)

	errorIfScriptFail(t, L, `
    ch = channel.make(10)
    `)
	obj = L.GetGlobal("ch")
	ch, _ = obj.(LChannel)
	errorIfNotEqual(t, 10, reflect.ValueOf(ch).Cap())
	close(ch)
}

func TestChannelSelectError(t *testing.T) {
	L := NewState()
	defer L.Close()
	errorIfScriptFail(t, L, `ch = channel.make()`)
	errorIfScriptNotFail(t, L, `channel.select({1,2,3})`, "invalid select case")
	errorIfScriptNotFail(t, L, `channel.select({"<-|", 1, 3})`, "invalid select case")
	errorIfScriptNotFail(t, L, `channel.select({"<-|", ch, function() end})`, "can not send a function")
	errorIfScriptNotFail(t, L, `channel.select({"|<-", 1, 3})`, "invalid select case")
	errorIfScriptNotFail(t, L, `channel.select({"<-->", 1, 3})`, "invalid channel direction")
	errorIfScriptFail(t, L, `ch:close()`)
}

func TestChannelSelect1(t *testing.T) {
	var result LValue
	var wg sync.WaitGroup
	receiver := func(ch, quit chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		L.SetGlobal("quit", LChannel(quit))
		if err := L.DoString(`
    buf = ""
    local exit = false
    while not exit do
      channel.select(
        {"|<-", ch, function(ok, v)
          if not ok then
            buf = buf .. "channel closed"
            exit = true
          else
            buf = buf .. "received:" .. v
          end
        end},
        {"|<-", quit, function(ok, v)
            buf = buf .. "quit"
        end}
      )
    end
  `); err != nil {
			panic(err)
		}
		result = L.GetGlobal("buf")
	}

	sender := func(ch, quit chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		L.SetGlobal("quit", LChannel(quit))
		if err := L.DoString(`
    ch:send("1")
    ch:send("2")
  `); err != nil {
			panic(err)
		}
		ch <- LString("3")
		quit <- LTrue
		time.Sleep(1 * time.Second)
		close(ch)
	}

	ch := make(chan LValue)
	quit := make(chan LValue)
	wg.Add(2)
	go receiver(ch, quit)
	go sender(ch, quit)
	wg.Wait()
	lstr, ok := result.(LString)
	errorIfFalse(t, ok, "must be string")
	str := string(lstr)
	errorIfNotEqual(t, "received:1received:2received:3quitchannel closed", str)

}

func TestChannelSelect2(t *testing.T) {
	var wg sync.WaitGroup
	receiver := func(ch, quit chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		L.SetGlobal("quit", LChannel(quit))
		errorIfScriptFail(t, L, `
           idx, rcv, ok = channel.select(
               {"|<-", ch},
               {"|<-", quit}
           )
           assert(idx == 1)
           assert(rcv == "1")
           assert(ok)
           idx, rcv, ok = channel.select(
               {"|<-", ch},
               {"|<-", quit}
           )
           assert(idx == 1)
           assert(rcv == nil)
           assert(not ok)
       `)
	}

	sender := func(ch, quit chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		L.SetGlobal("quit", LChannel(quit))
		errorIfScriptFail(t, L, `ch:send("1")`)
		errorIfScriptFail(t, L, `ch:close()`)
	}

	ch := make(chan LValue)
	quit := make(chan LValue)
	wg.Add(2)
	go receiver(ch, quit)
	go sender(ch, quit)
	wg.Wait()
}

func TestChannelSelect3(t *testing.T) {
	var wg sync.WaitGroup
	receiver := func(ch chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		errorIfScriptFail(t, L, `
           ok = true
           while ok do
             idx, rcv, ok = channel.select(
                 {"|<-", ch}
             )
           end
       `)
	}

	sender := func(ch chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		errorIfScriptFail(t, L, `
           ok = false
           channel.select(
               {"<-|", ch, "1", function(v)
                 ok = true
               end}
           )
           assert(ok)
           idx, rcv, ok = channel.select(
               {"<-|", ch, "1"}
           )
           assert(idx == 1)
           ch:close()
       `)
	}

	ch := make(chan LValue)
	wg.Add(2)
	go receiver(ch)
	time.Sleep(1)
	go sender(ch)
	wg.Wait()
}

func TestChannelSelect4(t *testing.T) {
	var wg sync.WaitGroup
	receiver := func(ch chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		errorIfScriptFail(t, L, `
           idx, rcv, ok = channel.select(
                 {"|<-", ch},
                 {"default"}
           )
           assert(idx == 2)
           called = false
           idx, rcv, ok = channel.select(
                 {"|<-", ch},
                 {"default", function()
                    called = true
                 end}
           )
           assert(called)
           ch:close()
       `)
	}

	ch := make(chan LValue)
	wg.Add(1)
	go receiver(ch)
	wg.Wait()
}

func TestChannelSendReceive1(t *testing.T) {
	var wg sync.WaitGroup
	receiver := func(ch chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		errorIfScriptFail(t, L, `
          local ok, v = ch:receive()
          assert(ok)
          assert(v == "1")
        `)
		time.Sleep(1 * time.Second)
		errorIfScriptFail(t, L, `
          local ok, v = ch:receive()
          assert(not ok)
          assert(v == nil)
        `)
	}
	sender := func(ch chan LValue) {
		defer wg.Done()
		L := NewState()
		defer L.Close()
		L.SetGlobal("ch", LChannel(ch))
		errorIfScriptFail(t, L, `ch:send("1")`)
		errorIfScriptNotFail(t, L, `ch:send(function() end)`, "can not send a function")
		errorIfScriptFail(t, L, `ch:close()`)
	}
	ch := make(chan LValue)
	wg.Add(2)
	go receiver(ch)
	go sender(ch)
	wg.Wait()
}
