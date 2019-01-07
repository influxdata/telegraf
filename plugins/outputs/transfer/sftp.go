package transfer

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

func (s *SftpTransferer) createConnection(host string, user string, pass string) (*sftp.Client, error) {
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

		return sftpConn, nil
	}

	return nil, err
}

func (s *SftpTransferer) getConnection(host string, user string, pass string) (*sftp.Client, error) {
	if conn, found := s.connections[host]; found {
		return conn, nil
	}

	// New connection
	conn, err := s.createConnection(host, user, pass)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *SftpTransferer) Send(source string, dest *url.URL) error {
	pass, _ := dest.User.Password()
	user := dest.User.Username()
	host := dest.Hostname()

	conn, err := s.getConnection(host, user, pass)
	if err != nil {
		return err
	}

	tmpFile := dest.Path + ".xtp"
	_, data, err := s.ReadFile(source)
	if err != nil {
		return err
	}

	// Create the destination file
	dstFile, err := conn.Create(tmpFile)
	if err != nil {
		// We could try to create the dest dir, but for now... just throw the file away
		log.Printf("ERROR [sftp.create] [%s]: %s", tmpFile, err)
		return err
	}

	// Move the file
	_, err = dstFile.Write(data)
	if err != nil {
		log.Println("ERROR [sftp.write]: ", err)
		return err
	}
	dstFile.Close()

	// Rename the file if we are using a temporary extension
	err = conn.Rename(tmpFile, dest.Path)
	if err != nil {
		log.Println("ERROR [sftp.rename]: ", err)
		return err
	}

	return nil
}
