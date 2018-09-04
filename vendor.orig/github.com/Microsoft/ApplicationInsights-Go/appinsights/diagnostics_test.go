package appinsights

import (
	"fmt"
	"testing"
	"time"
)

func TestMessageSentToConsumers(t *testing.T) {
	original := "~~~test_message~~~"

	// There may be spurious messages sent by a transmitter's goroutine from another test,
	// so just check that we do get the test message *at some point*.

	listener1chan := make(chan bool, 1)
	NewDiagnosticsMessageListener(func(message string) error {
		if message == original {
			listener1chan <- true
		}

		return nil
	})

	listener2chan := make(chan bool, 1)
	NewDiagnosticsMessageListener(func(message string) error {
		if message == original {
			listener2chan <- true
		}

		return nil
	})

	defer resetDiagnosticsListeners()
	diagnosticsWriter.Write(original)

	listener1recvd := false
	listener2recvd := false
	timeout := false
	timer := time.After(time.Second)
	for !(listener1recvd && listener2recvd) && !timeout {
		select {
		case <-listener1chan:
			listener1recvd = true
		case <-listener2chan:
			listener2recvd = true
		case <-timer:
			timeout = true
		}
	}

	if timeout {
		t.Errorf("Message failed to be delivered to both listeners")
	}
}

func TestRemoveListener(t *testing.T) {
	mchan := make(chan string, 1)
	listener := NewDiagnosticsMessageListener(func(message string) error {
		mchan <- message
		return nil
	})

	defer resetDiagnosticsListeners()

	diagnosticsWriter.Write("Hello")
	select {
	case <-mchan:
	default:
		t.Fatalf("Message not received")
	}

	listener.Remove()

	diagnosticsWriter.Write("Hello")
	select {
	case <-mchan:
		t.Fatalf("Message received after remove")
	default:
	}
}

func TestErroredListenerIsRemoved(t *testing.T) {
	mchan := make(chan string, 1)
	echan := make(chan error, 1)
	NewDiagnosticsMessageListener(func(message string) error {
		mchan <- message
		return <-echan
	})
	defer resetDiagnosticsListeners()

	echan <- nil
	diagnosticsWriter.Write("Hello")
	select {
	case <-mchan:
	default:
		t.Fatal("Message not received")
	}

	echan <- fmt.Errorf("Test error")
	diagnosticsWriter.Write("Hello")
	select {
	case <-mchan:
	default:
		t.Fatal("Message not received")
	}

	echan <- nil
	diagnosticsWriter.Write("Not received")
	select {
	case <-mchan:
		t.Fatalf("Message received after error")
	default:
	}
}

func resetDiagnosticsListeners() {
	diagnosticsWriter.lock.Lock()
	defer diagnosticsWriter.lock.Unlock()
	diagnosticsWriter.listeners = diagnosticsWriter.listeners[:0]
}
