package transferer

import (
	"bufio"
	"net/url"
	"os"
)

type Transferer interface {
	//Connect() error
	//Close() error
	Send(source string, dest *url.URL) error
	Rename(from *url.URL, to string) error
	//New(source string, dest string) *Transferer
}

type Credential struct {
	user string
	pass string
}

type CredentialMap struct {
	creds map[string]Credential
}

func (c *CredentialMap) Init() {
	c.creds = make(map[string]Credential)
}

func (c *CredentialMap) Get(host string) Credential {
	return c.creds[host]
}

func (c *CredentialMap) Add(host string, user string, pass string) {
	if c.creds == nil {
		c.Init()
	}

	cr := Credential{
		user: user,
		pass: pass,
	}

	c.creds[host] = cr
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
