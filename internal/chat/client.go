package chat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gasuhwbab/chat_server/internal/config"
	"github.com/gasuhwbab/chat_server/internal/proto"
)

type Client struct {
	conn    net.Conn
	address string
	nick    string
	out     chan string
	closed  atomic.Bool
	h       *Hub
	cfg     *config.Config
	id      int64
}

var clientSeq int64

func NewClient(h *Hub, conn net.Conn, cfg *config.Config) *Client {
	id := atomic.AddInt64(&clientSeq, 1)
	nick := fmt.Sprintf("user-%d", id)
	return &Client{
		conn:    conn,
		address: conn.RemoteAddr().String(),
		nick:    nick,
		out:     make(chan string, 64),
		h:       h,
		cfg:     cfg,
		id:      id,
	}
}

func (client *Client) CloseWithReason(reason string) {
	if client.closed.Swap(true) {
		return
	}
	client.conn.Close()
	select {
	case client.h.unregister <- client:
	default:
	}
}

func (client *Client) readLoop() {
	defer client.CloseWithReason("read loop exit")

	client.out <- "You are in go chat. Use /nick <name>, /who, /msg <nick> <text>, /help, /quit."
	client.out <- fmt.Sprintf("You are %s", client.nick)

	scanner := bufio.NewScanner(client.conn)
	buf := make([]byte, 0, client.cfg.MsgMaxBytes)
	scanner.Buffer(buf, client.cfg.MsgMaxBytes)
	for {
		client.conn.SetReadDeadline(time.Now().Add(client.cfg.IdleTimeout))
		if !scanner.Scan() {
			return
		}
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if len(line) == 0 {
			continue
		}
		if len(line) > client.cfg.MsgMaxBytes {
			client.out <- "Message is too long"
		}

		if proto.IsCommand(line) {
			client.handle(line)
			continue
		}
		client.h.broadcast <- proto.FormatUserMessage(client.nick, line)
	}
}

func (client *Client) handle(command string) {
	cmd := proto.Parse(command)
	switch cmd.Name {
	case "help":
		client.out <- "Use commands: /nick <name>, /who, /msg <nick> <text>, /quit"
	case "nick":
		if len(cmd.Args) < 1 {
			client.out <- "Error usage: /nick <name>"
			return
		}
		newNick := cmd.Args[0]
		resp := make(chan error, 1)
		client.h.rename <- renameReq{
			client:  client,
			newNick: newNick,
			resp:    resp,
		}
		if err := <-resp; err != nil {
			client.out <- err.Error()
		} else {
			client.out <- fmt.Sprintf("Your nick changed to %s", newNick)
		}
	case "who":
		req := whoReq{resp: make(chan []string, 1)}
		client.h.who <- req
		list := <-req.resp
		client.out <- "Online: " + strings.Join(list, ", ")
	case "msg":
		if len(cmd.Args) < 2 {
			client.out <- "Error usage: /msg <nick> <text>"
			return
		}
		nick := cmd.Args[0]
		text := strings.TrimSpace(strings.TrimPrefix(cmd.Raw, "/msg "+nick))
		if text == "" {
			client.out <- "Error empty message"
			return
		}
		client.h.priv <- PrivateMsg{from: client, to: nick, text: text}
	case "quit":
		client.out <- "Bye"
		client.CloseWithReason("User quit")
		return
	default:
		client.out <- "Unknown command. Try /help"
	}
}
func (client *Client) writeLoop() {
	defer client.CloseWithReason("write loop exit")
	w := bufio.NewWriter(client.conn)
	for {
		select {
		case msg, ok := <-client.out:
			if !ok {
				return
			}
			client.conn.SetWriteDeadline(time.Now().Add(client.cfg.WriteTimeout))
			if _, err := w.WriteString(msg + "\n"); err != nil {
				return
			}
			if err := w.Flush(); err != nil {
				return
			}
		}
	}
}
