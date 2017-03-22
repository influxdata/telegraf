package journalparser

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command

type journalClient struct {
	matches []string

	// matchPairs is a key/values mapping of matches.
	// The values are a disjunction of acceptable field values.
	matchPairs map[string][][]byte

	jeChan chan *journalEntry
}

type journalEntry struct {
	time   time.Time
	fields map[string][]byte
}

type tmpError struct{ error }

type journalStreamer struct {
	path string

	clients   []*journalClient
	cmd       *exec.Cmd
	reader    *bufio.Reader
	lastEntry *journalEntry
	wg        sync.WaitGroup
	sync.Mutex
}

var journalStreamers = map[string]*journalStreamer{}
var journalStreamersMtx sync.Mutex

func GetJournalStreamer(path string) *journalStreamer {
	journalStreamersMtx.Lock()
	js := journalStreamers[path]
	if js == nil {
		js = &journalStreamer{path: path}
	}
	journalStreamersMtx.Unlock()
	return js
}

func (js *journalStreamer) NewClient(matches []string) (*journalClient, error) {
	client, err := newJournalClient(matches)
	if err != nil {
		return nil, err
	}

	js.clients = append(js.clients, client)

	if err := js.start(); err != nil {
		return nil, err
	}

	return client, nil
}

func (js *journalStreamer) RemoveClient(rclient *journalClient) {
	var clients []*journalClient
	for _, client := range clients {
		if client != rclient {
			clients = append(clients, client)
		}
	}

	if len(clients) == len(js.clients) {
		return
	}
	js.clients = clients

	if len(clients) == 0 {
		js.stopWait()
	} else {
		// Even though we may fail to start up a new journalctl (extremely unlikely),
		// we can still close the chan, so ignore error.
		js.start()
	}

	close(rclient.jeChan)
}

func newJournalClient(matches []string) (*journalClient, error) {
	client := &journalClient{
		matches:    matches,
		matchPairs: map[string][][]byte{},
		jeChan:     make(chan *journalEntry, 1000),
	}
	for _, match := range matches {
		pair := strings.SplitN(match, "=", 2)
		if len(pair) != 2 || len(pair[0]) == 0 || pair[0][0] == '-' {
			return nil, fmt.Errorf("invalid match %q", match)
		}
		client.matchPairs[pair[0]] = append(client.matchPairs[pair[0]], []byte(pair[1]))
	}
	return client, nil
}

func (js *journalStreamer) start() error {
	if len(js.clients) == 0 {
		return nil
	}

	cmdArgs := []string{"-o", "export", "-f", "-n", "0"}

	if js.path == "user" {
		cmdArgs = append(cmdArgs, "--user")
	} else if js.path != "" {
		fi, err := os.Stat(js.path)
		if err != nil {
			return fmt.Errorf("unable to stat %s: %s", js.path, err)
		}
		if fi.IsDir() {
			cmdArgs = append(cmdArgs, "--directory", js.path)
		} else {
			cmdArgs = append(cmdArgs, "--file", js.path)
		}
	}

	matchArgs := []string{}
	for _, client := range js.clients {
		if len(matchArgs) != 0 {
			matchArgs = append(matchArgs, "+")
		}
		matchArgs = append(matchArgs, client.matches...)
	}
	cmd := execCommand("journalctl", append(cmdArgs, matchArgs...)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(stdout)
	if err := cmd.Start(); err != nil {
		return err
	}

	// ok, new command has started successfully, shut down and swap out the old one

	if js.cmd != nil {
		js.stopWait()
	}

	js.cmd = cmd
	js.reader = reader

	js.wg.Add(1)
	if js.lastEntry != nil {
		go js.run(&js.lastEntry.time)
	} else {
		go js.run(nil)
	}

	return nil
}

func (js *journalStreamer) run(after *time.Time) {
	defer js.wg.Done()
	for {
		je, err := js.readEntry()
		if err != nil {
			if _, ok := err.(tmpError); ok {
				log.Printf("error while reading journal entry: %s", err)
				continue
			}
			log.Printf("fatal error reading journal: %s", err)
			js.stop()
			//TODO restart with rate limit or something?
			return
		}

		if after != nil {
			if !je.time.After(*after) {
				continue
			}
			after = nil
		}

		js.dispatch(je)
		js.lastEntry = je
	}
}

func (js *journalStreamer) dispatch(je *journalEntry) {
	// iterate through the clients and send event to ones where all matches are found
CLIENT:
	for _, client := range js.clients {
		for mk, mvs := range client.matchPairs {
			ev, ok := je.fields[mk]
			if !ok {
				continue CLIENT
			}

			ok = false
			for _, mv := range mvs {
				if bytes.Equal(ev, mv) {
					ok = true
					break
				}
			}
			if !ok {
				continue CLIENT
			}
		}

		// all matches found
		select {
		case client.jeChan <- je:
		default:
			log.Printf("journal entry dropped") //TODO some sort of identifier to know who is misbehaving
		}
	}
}

func (js *journalStreamer) readEntry() (*journalEntry, error) {
	je := &journalEntry{
		fields: make(map[string][]byte),
	}

	for {
		k, v, err := js.readField()
		if err != nil {
			return nil, err
		}
		if k == "" {
			break
		}
		je.fields[k] = v
	}

	tsMicro, err := strconv.ParseUint(string(je.fields["__REALTIME_TIMESTAMP"]), 10, 64)
	if err != nil {
		return nil, tmpError{fmt.Errorf("unable to parse timestamp: %s", err)}
	}
	je.time = time.Unix(0, int64(time.Microsecond)*int64(tsMicro))

	return je, nil
}

// readField reads a field from the given reader.
// Returns the field's key, value, and any error.
// Empty key with no error denotes end of entry.
func (js *journalStreamer) readField() (string, []byte, error) {
	line, err := js.reader.ReadBytes('\n')
	if err != nil {
		return "", nil, err
	}
	if len(line) == 1 { // empty line denotes end of entry
		return "", nil, nil
	}
	line = line[:len(line)-1]

	var key string
	var value []byte
	split := bytes.SplitN(line, []byte{'='}, 2)
	key = string(split[0])
	if len(split) == 2 {
		value = split[1]
	} else {
		// field has a binary value
		var size uint64
		if err := binary.Read(js.reader, binary.LittleEndian, &size); err != nil {
			return "", nil, err
		}
		value = make([]byte, size)
		if _, err := js.reader.Read(value); err != nil {
			return "", nil, err
		}
		if _, err := js.reader.Discard(1); err != nil { // discards newline
			return "", nil, err
		}
	}

	return key, value, nil
}

func (js *journalStreamer) stop() {
	js.Lock()
	defer js.Unlock()
	if js.cmd == nil || js.cmd.Process == nil || js.cmd.ProcessState != nil {
		return
	}

	proc := js.cmd.Process
	proc.Signal(syscall.SIGTERM)
	timer := time.AfterFunc(time.Second, func() {
		proc.Kill()
	})
	_, err := proc.Wait()
	timer.Stop()
	if err != nil {
		proc.Kill()
	}
	js.cmd = nil
}
func (js *journalStreamer) stopWait() {
	js.stop()
	js.wg.Wait()
}
