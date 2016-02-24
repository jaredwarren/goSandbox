package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

const (
	PUBLIC_KEY = "C:/Users/jaredwarren/.ssh/id_rsa"
	IP         = "162.242.190.44"
	PORT       = "22"
)

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return ssh.PublicKeys(key)
}

func ExecSSH(command string) (err error) {
	fmt.Println("> ssh " + command)
	sshConfig := &ssh.ClientConfig{
		User: "jwarren",
		Auth: []ssh.AuthMethod{
			PublicKeyFile(PUBLIC_KEY),
		},
	}
	connection, err := ssh.Dial("tcp", IP+":"+PORT, sshConfig)
	if err != nil {
		fmt.Errorf("Failed to dial: %s", err)
	}

	session, err := connection.NewSession()
	if err != nil {
		fmt.Errorf("Failed to create session: %s", err)
	}

	err = session.Run(command)
	return
}
