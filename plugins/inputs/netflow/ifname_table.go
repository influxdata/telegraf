package netflow

import (
	"log"
)

type IfnameWriteOp struct {
	Key   uint32
	Value string
	Resp  chan bool
}

type IfnameReadOp struct {
	Key  uint32
	Resp chan string
	Fail chan bool
}

func (n *Netflow) ifnamePoller() error {
	defer n.wg.Done()
	var ifnameTable = map[uint32]string{}
	for {
		select {
		case <-n.done:
			return nil
		case read := <-n.readIfname:
			log.Println("D! read ifname")
			ifname, ok := ifnameTable[read.Key]
			if ok {
				read.Resp <- ifname
			} else {
				read.Fail <- false
			}
		case write := <-n.writeIfname:
			log.Println("D! write ifname")
			ifnameTable[write.Key] = write.Value
			write.Resp <- true
		}
	}
}
