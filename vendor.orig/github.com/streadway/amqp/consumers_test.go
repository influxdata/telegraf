package amqp

import (
	"strings"
	"testing"
)

func TestGeneratedUniqueConsumerTagDoesNotExceedMaxLength(t *testing.T) {
	assertCorrectLength := func(commandName string) {
		tag := commandNameBasedUniqueConsumerTag(commandName)
		if len(tag) > consumerTagLengthMax {
			t.Error("Generated unique consumer tag exceeds maximum length:", tag)
		}
	}

	assertCorrectLength("test")
	assertCorrectLength(strings.Repeat("z", 249))
	assertCorrectLength(strings.Repeat("z", 256))
	assertCorrectLength(strings.Repeat("z", 1024))
}
