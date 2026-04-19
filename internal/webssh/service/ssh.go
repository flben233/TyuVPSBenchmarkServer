package service

import (
	"VPSBenchmarkBackend/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"VPSBenchmarkBackend/internal/webssh/model"

	"golang.org/x/crypto/ssh"
)

var sessions = sync.Map{} // map[int64]*SSHSession

func GetSession(sessionID string) (*SSHSession, bool) {
	if sess, ok := sessions.Load(sessionID); ok {
		return sess.(*SSHSession), true
	}
	return nil, false
}

type sideReader struct {
	Buf       chan []byte
	SideClose chan struct{}
}

func (r *sideReader) Read(p []byte) (n int, err error) {
	data, ok := <-r.Buf
	if !ok {
		return 0, io.EOF
	}
	n = copy(p, data)
	return n, nil
}

type SSHSession struct {
	ID        string
	client    *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	mu        sync.Mutex
	done      chan struct{}
	closeOnce sync.Once
	sideOut   *sideReader
	sideMu    sync.Mutex
}

func NewSSHSession() *SSHSession {
	return &SSHSession{
		ID:   uuid.New().String(),
		done: make(chan struct{}),
	}
}

func (s *SSHSession) SetSideBuffer() io.Reader {
	buf := make(chan []byte)
	s.sideMu.Lock()
	defer s.sideMu.Unlock()
	s.sideOut = &sideReader{Buf: buf, SideClose: make(chan struct{})}
	return s.sideOut
}

func (s *SSHSession) ClearSideBuffer() {
	close(s.sideOut.SideClose) // 发送关闭信号，不要用写入来发送信号，避免下面的select阻塞（两个都是无缓冲通道，如果sideOut.Buf没有被读取，写入会阻塞）
	s.sideMu.Lock()
	defer s.sideMu.Unlock()
	if s.sideOut != nil {
		close(s.sideOut.Buf)
		s.sideOut = nil
	}
}

func (s *SSHSession) Connect(msg *model.ClientMessage) error {
	var authMethods []ssh.AuthMethod

	if msg.Password != "" {
		authMethods = append(authMethods, ssh.Password(msg.Password))
	}
	if msg.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(msg.PrivateKey))
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication method provided")
	}

	cfg := &ssh.ClientConfig{
		User:            msg.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", msg.Host, msg.Port)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create session: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	termRows := msg.Rows
	termCols := msg.Cols
	if termRows == 0 {
		termRows = 24
	}
	if termCols == 0 {
		termCols = 80
	}

	if err := session.RequestPty("xterm-256color", termRows, termCols, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	if err := session.Shell(); err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return fmt.Errorf("failed to start shell: %w", err)
	}

	s.client = client
	s.session = session
	s.stdin = stdin
	s.stdout = stdout
	sessions.Store(s.ID, s)
	return nil
}

func (s *SSHSession) Resize(rows, cols int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.session == nil {
		return fmt.Errorf("session not active")
	}
	return s.session.WindowChange(rows, cols)
}

func (s *SSHSession) WriteInput(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stdin == nil {
		return fmt.Errorf("session not active")
	}

	_, err := s.stdin.Write(data)
	return err
}

func (s *SSHSession) ReadOutput(sendOutput func([]byte), sendMsg func(*model.ServerMessage)) {
	buf := make([]byte, 8192)
	for {
		select {
		case <-s.done:
			return
		default:
		}

		n, err := s.stdout.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			sendOutput(data)
			// 如果侧边工具有在监听输出，也发送数据过去
			s.sideMu.Lock()
			if s.sideOut != nil {
				select {
				case <-s.sideOut.SideClose:
					fmt.Println("Side buffer closed, stop sending data")
					// 释放锁
				case s.sideOut.Buf <- data:
					fmt.Println("Sent data to side buffer")
				}
			}
			s.sideMu.Unlock()
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("SSH read error: %v", err)
			}
			s.closeOnce.Do(func() { close(s.done) })
			sendMsg(&model.ServerMessage{
				Type: model.TypeClosed,
			})
			return
		}
	}
}

func (s *SSHSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closeOnce.Do(func() { close(s.done) })

	if s.stdin != nil {
		s.stdin.Close()
	}
	if s.session != nil {
		s.session.Close()
	}
	if s.client != nil {
		s.client.Close()
	}

	// 通知LLM后端会话已关闭
	closeReq := &struct {
		SSHSessionID string `json:"ssh_session_id"`
	}{
		SSHSessionID: s.ID,
	}
	reqJson, err := json.Marshal(closeReq)
	if err == nil {
		_, _ = http.Post("POST", config.Get().LLMBackendURL+"/close", bytes.NewReader(reqJson))
	}

	sessions.Delete(s.ID)
}
