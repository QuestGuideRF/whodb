package ssh
import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)
type Tunnel struct {
	sshClient  *ssh.Client
	listener   net.Listener
	remoteHost string
	remotePort int
	done       chan struct{}
	closeOnce  sync.Once
}
func NewTunnel(sshHost string, sshPort int, sshUser, sshKeyFile, sshPassword string, remoteHost string, remotePort int) (*Tunnel, error) {
	authMethods, err := buildAuthMethods(sshKeyFile, sshPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH auth methods: %w", err)
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no SSH authentication method available: provide a key file, password, or ensure ssh-agent is running")
	}
	hostKeyCallback, err := defaultHostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH known hosts: %w", err)
	}
	sshConfig := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
	}
	sshAddr := fmt.Sprintf("%s:%d", sshHost, sshPort)
	client, err := ssh.Dial("tcp", sshAddr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server %s: %w", sshAddr, err)
	}
	return &Tunnel{
		sshClient:  client,
		remoteHost: remoteHost,
		remotePort: remotePort,
		done:       make(chan struct{}),
	}, nil
}
func (t *Tunnel) Start() error {
	var err error
	t.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to open local listener: %w", err)
	}
	go t.acceptLoop()
	return nil
}
func (t *Tunnel) Stop() {
	t.closeOnce.Do(func() {
		close(t.done)
		if t.listener != nil {
			t.listener.Close()
		}
		if t.sshClient != nil {
			t.sshClient.Close()
		}
	})
}
func (t *Tunnel) LocalPort() int {
	if t.listener == nil {
		return 0
	}
	return t.listener.Addr().(*net.TCPAddr).Port
}
func (t *Tunnel) acceptLoop() {
	remoteAddr := fmt.Sprintf("%s:%d", t.remoteHost, t.remotePort)
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.done:
				return
			default:
				continue
			}
		}
		go t.forward(localConn, remoteAddr)
	}
}
func (t *Tunnel) forward(localConn net.Conn, remoteAddr string) {
	remoteConn, err := t.sshClient.Dial("tcp", remoteAddr)
	if err != nil {
		localConn.Close()
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(remoteConn, localConn)
		remoteConn.Close()
	}()
	go func() {
		defer wg.Done()
		io.Copy(localConn, remoteConn)
		localConn.Close()
	}()
	wg.Wait()
}
func defaultHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("~/.ssh/known_hosts not found — run 'ssh %s' once to add the host key", "user@host")
	}
	return knownhosts.New(knownHostsPath)
}
func buildAuthMethods(keyFile, password string) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if keyFile != "" {
		keyData, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSH key file %q: %w", keyFile, err)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH key file %q: %w", keyFile, err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if authSock := os.Getenv("SSH_AUTH_SOCK"); authSock != "" {
		conn, err := net.Dial("unix", authSock)
		if err == nil {
			agentClient := agent.NewClient(conn)
			methods = append(methods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}
	if password != "" {
		methods = append(methods, ssh.Password(password))
	}
	return methods, nil
}