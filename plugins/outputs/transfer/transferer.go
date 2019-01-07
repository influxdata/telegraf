package transfer

import (
	"bufio"
	"net/url"
	"os"
)

type Transferer interface {
	//Connect() error
	//Close() error
	Send(source string, dest *url.URL) error
	//New(source string, dest string) *Transferer
}

type BaseTransferer struct{}

func (t *BaseTransferer) ReadFile(filename string) (int, []byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return 0, nil, statsErr
	}

	var size int = int(stats.Size())
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return size, bytes, err
}
