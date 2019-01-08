package nsq_consumer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
)

// This test is modeled after the kafka consumer integration test
func TestReadsMetricsFromNSQ(t *testing.T) {
	msgID := nsq.MessageID{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0', 'a', 's', 'd', 'f', 'g', 'h'}
	msg := nsq.NewMessage(msgID, []byte("cpu_load_short,direction=in,host=server01,region=us-west value=23422.0 1422568543702900257\n"))

	script := []instruction{
		// SUB
		{0, nsq.FrameTypeResponse, []byte("OK")},
		// IDENTIFY
		{0, nsq.FrameTypeResponse, []byte("OK")},
		{20 * time.Millisecond, nsq.FrameTypeMessage, frameMessage(msg)},
		// needed to exit test
		{100 * time.Millisecond, -1, []byte("exit")},
	}

	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:4155")
	newMockNSQD(script, addr.String())

	consumer := &NSQConsumer{
		Server:                 "127.0.0.1:4155",
		Topic:                  "telegraf",
		Channel:                "consume",
		MaxInFlight:            1,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Nsqd:                   []string{"127.0.0.1:4155"},
	}

	p, _ := parsers.NewInfluxParser()
	consumer.SetParser(p)
	var acc testutil.Accumulator
	assert.Equal(t, 0, len(acc.Metrics), "There should not be any points")
	if err := consumer.Start(&acc); err != nil {
		t.Fatal(err.Error())
	} else {
		defer consumer.Stop()
	}

	waitForPoint(&acc, t)

	if len(acc.Metrics) == 1 {
		point := acc.Metrics[0]
		assert.Equal(t, "cpu_load_short", point.Measurement)
		assert.Equal(t, map[string]interface{}{"value": 23422.0}, point.Fields)
		assert.Equal(t, map[string]string{
			"host":      "server01",
			"direction": "in",
			"region":    "us-west",
		}, point.Tags)
		assert.Equal(t, time.Unix(0, 1422568543702900257).Unix(), point.Time.Unix())
	} else {
		t.Errorf("No points found in accumulator, expected 1")
	}

}

// Waits for the metric that was sent to the kafka broker to arrive at the kafka
// consumer
func waitForPoint(acc *testutil.Accumulator, t *testing.T) {
	// Give the kafka container up to 2 seconds to get the point to the consumer
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			if counter > 1000 {
				t.Fatal("Waited for 5s, point never arrived to consumer")
			} else if acc.NFields() == 1 {
				return
			}
		}
	}
}

func newMockNSQD(script []instruction, addr string) *mockNSQD {
	n := &mockNSQD{
		script:   script,
		exitChan: make(chan int),
	}

	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("FATAL: listen (%s) failed - %s", n.tcpAddr.String(), err)
	}
	n.tcpListener = tcpListener
	n.tcpAddr = tcpListener.Addr().(*net.TCPAddr)

	go n.listen()

	return n
}

// The code below allows us to mock the interactions with nsqd. This is taken from:
// https://github.com/nsqio/go-nsq/blob/master/mock_test.go
type instruction struct {
	delay     time.Duration
	frameType int32
	body      []byte
}

type mockNSQD struct {
	script      []instruction
	got         [][]byte
	tcpAddr     *net.TCPAddr
	tcpListener net.Listener
	exitChan    chan int
}

func (n *mockNSQD) listen() {
	for {
		conn, err := n.tcpListener.Accept()
		if err != nil {
			break
		}
		go n.handle(conn)
	}
	close(n.exitChan)
}

func (n *mockNSQD) handle(conn net.Conn) {
	var idx int
	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		log.Fatalf("ERROR: failed to read protocol version - %s", err)
	}

	readChan := make(chan []byte)
	readDoneChan := make(chan int)
	scriptTime := time.After(n.script[0].delay)
	rdr := bufio.NewReader(conn)

	go func() {
		for {
			line, err := rdr.ReadBytes('\n')
			if err != nil {
				return
			}
			// trim the '\n'
			line = line[:len(line)-1]
			readChan <- line
			<-readDoneChan
		}
	}()

	var rdyCount int
	for idx < len(n.script) {
		select {
		case line := <-readChan:
			n.got = append(n.got, line)
			params := bytes.Split(line, []byte(" "))
			switch {
			case bytes.Equal(params[0], []byte("IDENTIFY")):
				l := make([]byte, 4)
				_, err := io.ReadFull(rdr, l)
				if err != nil {
					log.Printf(err.Error())
					goto exit
				}
				size := int32(binary.BigEndian.Uint32(l))
				b := make([]byte, size)
				_, err = io.ReadFull(rdr, b)
				if err != nil {
					log.Printf(err.Error())
					goto exit
				}
			case bytes.Equal(params[0], []byte("RDY")):
				rdy, _ := strconv.Atoi(string(params[1]))
				rdyCount = rdy
			case bytes.Equal(params[0], []byte("FIN")):
			case bytes.Equal(params[0], []byte("REQ")):
			}
			readDoneChan <- 1
		case <-scriptTime:
			inst := n.script[idx]
			if bytes.Equal(inst.body, []byte("exit")) {
				goto exit
			}
			if inst.frameType == nsq.FrameTypeMessage {
				if rdyCount == 0 {
					scriptTime = time.After(n.script[idx+1].delay)
					continue
				}
				rdyCount--
			}
			_, err := conn.Write(framedResponse(inst.frameType, inst.body))
			if err != nil {
				log.Printf(err.Error())
				goto exit
			}
			scriptTime = time.After(n.script[idx+1].delay)
			idx++
		}
	}

exit:
	n.tcpListener.Close()
	conn.Close()
}

func framedResponse(frameType int32, data []byte) []byte {
	var w bytes.Buffer

	beBuf := make([]byte, 4)
	size := uint32(len(data)) + 4

	binary.BigEndian.PutUint32(beBuf, size)
	_, err := w.Write(beBuf)
	if err != nil {
		return nil
	}

	binary.BigEndian.PutUint32(beBuf, uint32(frameType))
	_, err = w.Write(beBuf)
	if err != nil {
		return nil
	}

	w.Write(data)
	return w.Bytes()
}

func frameMessage(m *nsq.Message) []byte {
	var b bytes.Buffer
	m.WriteTo(&b)
	return b.Bytes()
}
