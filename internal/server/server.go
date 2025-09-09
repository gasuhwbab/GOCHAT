package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gasuhwbab/chat_server/internal/chat"
	"github.com/gasuhwbab/chat_server/internal/config"
	"github.com/gasuhwbab/chat_server/internal/proto"
	"github.com/gasuhwbab/chat_server/internal/transport/tcp"
)

type Server struct {
	cfg config.Config
	hub *chat.Hub
	ln  *tcp.Listener
}

func NewServer(cfg config.Config) *Server {
	return &Server{
		cfg: cfg,
		hub: chat.NewHub(),
	}
}

func (serv *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", serv.cfg.Host, serv.cfg.Port)
	ln, err := tcp.NewListener(addr)
	if err != nil {
		return err
	}
	serv.ln = ln
	log.Println("listening on:", serv.ln.Addr())

	doneHub := make(chan struct{})
	go func() {
		serv.hub.StartLoop()
		close(doneHub)
	}()

	acceptErr := make(chan error, 1)
	go func() {
		for {
			conn, err := serv.ln.Accept()
			if err != nil {
				if _, ok := err.(net.Error); ok {
					log.Println("Network error")
					time.Sleep(100 * time.Millisecond)
					continue
				}
				acceptErr <- err
				return
			}

			if serv.cfg.MaxClients > 0 && serv.hub.ClientsCurrent() >= int64(serv.cfg.MaxClients) {
				conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
				conn.Write([]byte("Error server is full, please try again later"))
				conn.Close()
				continue
			}
			c := chat.NewClient(serv.hub, conn, &serv.cfg)
			serv.hub.Register(c)
		}
	}()

	select {
	case <-ctx.Done():
		serv.hub.Broadcast(proto.FormatSystemMessage("SYSTEM SHOTDOWN IN 3 SECONDS"))
		time.Sleep(3 * time.Second)
		serv.hub.CloseAll()
		serv.ln.Close()
		<-doneHub
		return nil
	case err := <-acceptErr:
		serv.ln.Close()
		serv.hub.CloseAll()
		<-doneHub
		return err
	}
}
