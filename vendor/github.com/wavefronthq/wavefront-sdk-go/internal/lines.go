package internal

import (
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"
)

type LineHandler struct {
	Reporter      Reporter
	BatchSize     int
	MaxBufferSize int
	FlushTicker   *time.Ticker
	Format        string
	failures      int64
	buffer        chan string
	done          chan bool
}

func (lh *LineHandler) Start() {
	lh.buffer = make(chan string, lh.MaxBufferSize)
	lh.done = make(chan bool)

	go func() {
		for {
			select {
			case <-lh.FlushTicker.C:
				err := lh.Flush()
				if err != nil {
					log.Println(err)
				}
			case <-lh.done:
				return
			}
		}
	}()
}

func (lh *LineHandler) HandleLine(line string) error {
	select {
	case lh.buffer <- line:
		return nil
	default:
		atomic.AddInt64(&lh.failures, 1)
		return fmt.Errorf("buffer full, dropping line: %s", line)
	}
}

func (lh *LineHandler) Flush() error {
	bufLen := len(lh.buffer)
	if bufLen > 0 {
		size := min(bufLen, lh.BatchSize)
		lines := make([]string, size)
		for i := 0; i < size; i++ {
			lines[i] = <-lh.buffer
		}
		return lh.report(lines)
	}
	return nil
}

func (lh *LineHandler) report(lines []string) error {
	strLines := strings.Join(lines, "")
	resp, err := lh.Reporter.Report(lh.Format, strLines)

	if err != nil || (400 <= resp.StatusCode && resp.StatusCode <= 599) {
		atomic.AddInt64(&lh.failures, 1)
		lh.bufferLines(lines)
		if err != nil {
			return fmt.Errorf("error reporting %s format data to Wavefront: %q", lh.Format, err)
		} else {
			return fmt.Errorf("error reporting %s format data to Wavefront. status=%d", lh.Format, resp.StatusCode)
		}
	}
	return nil
}

func (lh *LineHandler) bufferLines(batch []string) {
	log.Println("error reporting to Wavefront. buffering lines.")
	for _, line := range batch {
		lh.HandleLine(line)
	}
}

func (lh *LineHandler) GetFailureCount() int64 {
	return atomic.LoadInt64(&lh.failures)
}

func (lh *LineHandler) Stop() {
	lh.Flush()
	close(lh.done)
	lh.FlushTicker.Stop()
}
