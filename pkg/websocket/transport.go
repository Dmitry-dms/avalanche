package websocket

import (
	"fmt"
	"io/ioutil"
	"net"
	//"sync"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
)

type CustomWebsocket struct {
	//mu      sync.RWMutex
	closed  bool
	conn    net.Conn
}


type Websocket interface {
	Write(msg []byte) error
	Read() ([]byte, bool, error)
	Close() error
}


func NewWebsocketTransport(conn net.Conn) CustomWebsocket {
	return CustomWebsocket{
		conn:    conn,
	}
}
func (t *CustomWebsocket) IsClosed() bool {
	return t.closed
}



func (t CustomWebsocket) Read() ([]byte, bool, error) {
	//t.mu.Lock()
	//defer t.mu.Unlock()

	h, r, err := wsutil.NextReader(t.conn, ws.StateServerSide)
	if err != nil {
		return nil, true, errors.Wrap(err, "reader error")
	}

	//TODO: make right header parsing between gorilla and gobwas (opCode is missing in gorilla)
	if h.OpCode == ws.OpPing {
		fmt.Println("called ping, sending pong")
		return nil, false, wsutil.WriteServerMessage(t.conn, ws.OpPong, []byte{})
	}
	if h.OpCode == ws.OpPong {
		fmt.Println("called pong")
		return nil, false, nil
	}

	if h.OpCode.IsControl() {
		return nil, true, wsutil.ControlFrameHandler(t.conn, ws.StateServerSide)(h, r)
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, false, err
	}

	return data, false, nil
}

func (t CustomWebsocket) Write(msg []byte) error {
	if err := wsutil.WriteServerMessage(t.conn, ws.OpText, msg); err != nil {
		return err
	}
	return nil
}

func (t CustomWebsocket) Close() error {
	// if t.closed {
	// 	return nil
	// }
	// t.closed = true
	//close(t.closeCh)
	data := ws.NewCloseFrameBody(ws.StatusNormalClosure, "closing connection")
	_ = wsutil.WriteServerMessage(t.conn, ws.OpClose, data)
	return t.conn.Close()
}
