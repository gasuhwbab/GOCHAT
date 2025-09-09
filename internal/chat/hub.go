package chat

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gasuhwbab/chat_server/internal/proto"
)

type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan string
	priv       chan PrivateMsg
	rename     chan renameReq
	who        chan whoReq

	clients map[*Client]bool
	byNick  map[string]*Client

	clientsCurrent int64
	closeAll       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan string, 1024),
		priv:       make(chan PrivateMsg, 256),
		rename:     make(chan renameReq),
		who:        make(chan whoReq),

		clients: make(map[*Client]bool),
		byNick:  make(map[string]*Client),

		closeAll: make(chan struct{}),
	}
}

func (h *Hub) StartLoop() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			h.byNick[c.nick] = c
			atomic.AddInt64(&h.clientsCurrent, 1)
			h.systemBroadcast(fmt.Sprintf("* %s joined (%s)", c.nick, c.address))
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				if cur, okk := h.byNick[c.nick]; okk && cur == c {
					delete(h.byNick, c.nick)
				}
				atomic.AddInt64(&h.clientsCurrent, -1)
				h.systemBroadcast(fmt.Sprintf("* %s left", c.nick))
			}
		case s := <-h.broadcast:
			for cl := range h.clients {
				select {
				case cl.out <- s:
				default:
					go cl.CloseWithReason("slow client")
				}
			}
		case p := <-h.priv:
			if to, ok := h.byNick[p.to]; ok {
				msg := proto.FormatPrivateMessage(p.from.nick, p.to, p.text)
				select {
				case to.out <- msg:
				default:
					go to.CloseWithReason("slow client")
				}
				select {
				case p.from.out <- msg:
				default:
					go p.from.CloseWithReason("slow client")
				}
			} else {
				select {
				case p.from.out <- "No such nick " + p.to:
				default:
				}
			}
		case req := <-h.rename:
			if !proto.ValidNick(req.newNick) {
				req.resp <- fmt.Errorf("bad nick, use a-z, A-Z, 0-9")
				continue
			}
			if c, ok := h.byNick[req.newNick]; ok && req.client != c {
				req.resp <- fmt.Errorf("error: nick is used")
				continue
			}
			old := req.client.nick
			req.client.nick = req.newNick
			delete(h.byNick, old)
			h.byNick[req.client.nick] = req.client
			h.systemBroadcast(fmt.Sprintf("* %s changed nick to %s", old, req.client.nick))
			req.resp <- nil
		case req := <-h.who:
			list := make([]string, 0)
			for c := range h.clients {
				list = append(list, c.nick)
			}
			req.resp <- list
		case <-h.closeAll:
			for c := range h.clients {
				c.CloseWithReason("server shotdown")
			}
			return
		}
	}
}

func (h *Hub) systemBroadcast(text string) {
	stamp := time.Now().Format("15:04")
	h.broadcast <- fmt.Sprintf("[%s] %s", stamp, text)
}

func (h *Hub) ClientsCurrent() int64 {
	return atomic.LoadInt64(&h.clientsCurrent)
}

func (h *Hub) CloseAll() {
	select {
	case h.closeAll <- struct{}{}:
	default:
	}
}

func (h *Hub) Register(c *Client) {
	go c.writeLoop()
	go c.readLoop()
	h.register <- c
}
func (h *Hub) Broadcast(s string) { h.broadcast <- s }
