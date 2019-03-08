package logger

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteLogToFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()

	SetupLogging(false, false, tmpfile.Name())
	log.Printf("I! TEST")
	log.Printf("D! TEST") // <- should be ignored

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestDebugWriteLogToFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()

	SetupLogging(true, false, tmpfile.Name())
	log.Printf("D! TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z D! TEST\n"))
}

func TestErrorWriteLogToFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()

	SetupLogging(false, true, tmpfile.Name())
	log.Printf("E! TEST")
	log.Printf("I! TEST") // <- should be ignored

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z E! TEST\n"))
}

func TestAddDefaultLogLevel(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()

	SetupLogging(true, false, tmpfile.Name())
	log.Printf("TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))
}

func TestWriteToTruncatedFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() { os.Remove(tmpfile.Name()) }()

	SetupLogging(true, false, tmpfile.Name())
	log.Printf("TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! TEST\n"))

	tmpf, err := os.OpenFile(tmpfile.Name(), os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NoError(t, tmpf.Close())

	log.Printf("SHOULD BE FIRST")

	f, err = ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, f[19:], []byte("Z I! SHOULD BE FIRST\n"))
}

func BenchmarkTelegrafLogWrite(b *testing.B) {
	var msg = []byte("test")
	var buf bytes.Buffer
	w := newTelegrafWriter(&buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		w.Write(msg)
	}
}
