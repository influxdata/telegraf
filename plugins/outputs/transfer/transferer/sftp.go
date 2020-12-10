package transferer

import (
	"log"
	"net/url"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SftpTransferer struct {
	BaseTransferer

	connections map[string]*sftp.Client
}

func NewSftpTransferer() *SftpTransferer {
	return &SftpTransferer{
		connections: make(map[string]*sftp.Client),
	}
}

func (s *SftpTransferer) createConnection(url *url.URL) (*sftp.Client, error) {
	port := "22"
	if url.Port() != "" {
		port = url.Port()
	}
	host := url.Hostname() + ":" + port
	user := url.User.Username()
	pass, _ := url.User.Password()

	sshCommonConfig := ssh.Config{
		Ciphers: []string{
			"3des-cbc",
			"blowfish-cbc",
			"aes128-cbc",
			"aes128-ctr",
			"aes256-ctr",
		},
	}
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		Config:          sshCommonConfig,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var sshConn *ssh.Client
	var sftpConn *sftp.Client
	var err error
	for i := 0; i < 10; i++ {
		sshConn, err = ssh.Dial("tcp", host, sshConfig)
		if err != nil {
			log.Println("ERROR [ssh]: ", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}

		// Connect to the sftp server
		sftpConn, err = sftp.NewClient(sshConn)
		if err != nil {
			log.Println("ERROR [sftp]: ", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			continue
		}

		s.connections[url.Hostname()] = sftpConn
		return sftpConn, nil
	}

	return nil, err
}

func (s *SftpTransferer) getConnection(url *url.URL) (*sftp.Client, error) {
	if conn, found := s.connections[url.Hostname()]; found {
		return conn, nil
	}

	// New connection
	conn, err := s.createConnection(url)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *SftpTransferer) Rename(from *url.URL, to string) error {
	conn, err := s.getConnection(from)
	if err != nil {
		return err
	}

	err = conn.Rename(from.Path, to)
	if err != nil {
		log.Println("ERROR [sftp.rename]: ", err)
		return err
	}

	return nil
}

func (s *SftpTransferer) Send(source string, dest *url.URL) error {
	conn, err := s.getConnection(dest)
	if err != nil {
		return err
	}

	_, data, err := s.ReadFile(source)
	if err != nil {
		log.Printf("ERROR [sftp.read] [%s]: %s", source, err)
		return err
	}

	// Create the destination file
	dstFile, err := conn.Create(dest.Path)
	if err != nil {
		// We could try to create the dest dir, but for now... just throw the file away
		log.Printf("ERROR [sftp.create] [%s]: %s", dest.Path, err)
		return err
	}

	// Move the file
	_, err = dstFile.Write(data)
	if err != nil {
		log.Println("ERROR [sftp.write]: ", err)
		return err
	}
	dstFile.Close()

	return nil
}
