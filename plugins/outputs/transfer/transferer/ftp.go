package transferer

import (
	"log"
	"net/url"
	"os"

	"github.com/jlaffaye/ftp"
)

type FtpTransferer struct {
	BaseTransferer

	connections map[string]*ftp.ServerConn
}

func NewFtpTransferer() *FtpTransferer {
	return &FtpTransferer{
		connections: make(map[string]*ftp.ServerConn),
	}
}

func (f *FtpTransferer) createConnection(host string, user string, pass string) (*ftp.ServerConn, error) {
	conn, err := ftp.Connect(host)
	if err != nil {
		return nil, err
	}

	err = conn.Login(user, pass)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (f *FtpTransferer) getConnection(host string, user string, pass string) (*ftp.ServerConn, error) {
	if conn, found := f.connections[host]; found {
		return conn, nil
	}

	// New connection
	conn, err := f.createConnection(host, user, pass)
	if err != nil {
		return nil, err
	}

	f.connections[host] = conn
	return conn, nil
}

func (f *FtpTransferer) Rename(from *url.URL, to string) error {
	return nil
}

func (f *FtpTransferer) Send(source string, dest *url.URL) error {
	var err error
	var conn *ftp.ServerConn

	pwd, _ := dest.User.Password()
	user := dest.User.Username()
	port := "21"
	if dest.Port() != "" {
		port = dest.Port()
	}
	host := dest.Hostname() + ":" + port

	conn, err = f.getConnection(host, user, pwd)
	if err != nil {
		log.Printf("E! Could not connect to [%s]: %v", dest.Host, err)
		return err
	}

	r, err := os.Open(source)
	if err != nil {
		return err
	}
	defer r.Close()

	err = conn.StorFrom(dest.Path, r, 0)
	if err != nil {
		conn.Quit()
		conn = nil
		delete(f.connections, host)
		log.Printf("ERROR [ftp.storfrom] [%s]: %s", dest, err)
		return err
	}

	return nil
}
