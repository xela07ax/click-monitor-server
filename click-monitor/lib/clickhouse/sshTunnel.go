package clickhouse

// mysqlSSHtunnel project sshTunnel.go
import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/xela07ax/click-monitor-server/click-monitor/model"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
)

//
// Everything regarding the ssh tunnel goes here. Credits go to Svett Ralchev.
// Look at http://blog.ralch.com/tutorial/golang-ssh-tunneling for an excellent
// explanation and most ssh-tunneling related details used in this code.
//
// PEM key decryption is valid for password proected SSH-2 RSA Keys generated as
// .ppk files for putty and exported as OpenSSH .pem keyfile using PuTTYgen.
//
// Define an endpoint with ip and port
type Endpoint struct {
	Host string
	Port int
}

// Returns an endpoint as ip:port formatted string
func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

// Define the endpoints along the tunnel
type SSHtunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint
	Config *ssh.ClientConfig
}

// Start the tunnel
func (tunnel *SSHtunnel) Start() {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go tunnel.forward(conn)
	}
}

// Port forwarding
func (tunnel *SSHtunnel) forward(localConn net.Conn) {
	// Establish connection to the intermediate server
	// log.Printf("dialSsh:%s", tunnel.Server.String())
	tunnel.Config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		log.Fatalf("Server dial error: %s", err)
		return
	}

	// access the target server
	// log.Printf("dialServer:%s", tunnel.Remote.String())
	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		log.Fatalf("Remote dial error: %s", err)
		return
	}

	// Transfer the data between  and the remote server
	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			log.Fatalf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

// Decrypt encrypted PEM key data with a passphrase and embed it to key prefix
// and postfix header data to make it valid for further private key parsing.
func DecryptPEMkey(buffer []byte, passphrase string) []byte {
	block, _ := pem.Decode(buffer)
	der, err := x509.DecryptPEMBlock(block, []byte(passphrase))
	if err != nil {
		fmt.Println("decrypt failed: ", err)
	}
	encoded := base64.StdEncoding.EncodeToString(der)
	encoded = "-----BEGIN RSA PRIVATE KEY-----" + encoded + "-----END RSA PRIVATE KEY-----"
	return []byte(encoded)
}

// Get the signers from the OpenSSH key file (.pem) and return them for use in
// the Authentication method. Decrypt encrypted key data with the passphrase.
func PublicKeyFile(file string) ssh.AuthMethod {
	// A public key may be used to authenticate against the remote
	// server by using an unencrypted PEM-encoded private key file.
	//
	// If you have an encrypted private key, the crypto/x509 package
	// can be used to decrypt it.
	key, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	return ssh.PublicKeys(signer)
}

// Define the ssh tunnel using its endpoint and config data
func sshTunnel(sshConf *model.Ssh) *SSHtunnel {
	localEndpoint := &Endpoint{
		Host: sshConf.LocalHost,
		Port: sshConf.LocalPort,
	}

	serverEndpoint := &Endpoint{
		Host: sshConf.ServerHost,
		Port: sshConf.ServerPort,
	}

	remoteEndpoint := &Endpoint{
		Host: sshConf.RemoteHost,
		Port: sshConf.RemotePort,
	}

	sshConfig := &ssh.ClientConfig{
		User: sshConf.UserName,
		Auth: []ssh.AuthMethod{
			PublicKeyFile(sshConf.PrivateKeyFile)},
	}

	return &SSHtunnel{
		Config: sshConfig,
		Local:  localEndpoint,
		Server: serverEndpoint,
		Remote: remoteEndpoint,
	}
}
