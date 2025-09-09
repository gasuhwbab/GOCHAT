package tcp

import "net"

type Listener struct {
	ln   net.Listener
	addr string
}

func NewListener(addr string) (*Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Listener{ln: ln, addr: addr}, nil
}

func (l *Listener) Addr() string { return l.addr }

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (l *Listener) Close() error { return l.ln.Close() }
