package websocket

import (
	"fmt"
	"io/ioutil"
	//"io/ioutil"
	//"log"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const websocketTransportName = "websocket"

type CustomWebsocketTransport struct {
	mu      sync.RWMutex
	closed  bool
	closeCh chan struct{}
	conn    net.Conn
}

func NewWebsocketTransport(conn net.Conn) *CustomWebsocketTransport {
	return &CustomWebsocketTransport{
		conn:    conn,
		closeCh: make(chan struct{}),
	}
}
func (t *CustomWebsocketTransport) IsClosed() bool {
	return t.closed
}

func (t *CustomWebsocketTransport) CloseCh() chan struct{} {
	return t.closeCh
}

// Name implementation.
func (t *CustomWebsocketTransport) Name() string {
	return websocketTransportName
}

// Protocol implementation.
// func (t *customWebsocketTransport) Protocol() centrifuge.ProtocolType {
// 	return t.protoType
// }

// Encoding implementation.
// func (t *customWebsocketTransport) Encoding() centrifuge.EncodingType {
// 	return centrifuge.EncodingTypeJSON
// }

func (t *CustomWebsocketTransport) Read() ([]byte, bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	h, r, err := wsutil.NextReader(t.conn, ws.StateServerSide)
	if err != nil {
		return nil, false, err
	}
	if h.OpCode == ws.OpPing{
		return nil, true, wsutil.WriteServerMessage(t.conn, ws.OpPong, []byte("pong"))
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

func (t *CustomWebsocketTransport) Write(ch <-chan string ) error {
	select {
	case <-t.closeCh:
		return nil
	case msg:=<-ch:
		messageType := ws.OpBinary
		if err := wsutil.WriteServerMessage(t.conn, messageType, []byte(msg)); err != nil {
			return err
		}
		return nil
	}
}

// Close implementation.
func (t *CustomWebsocketTransport) Close() error {
	t.mu.Lock()
	fmt.Println("closing")
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	close(t.closeCh)
	t.mu.Unlock()
	// if disconnect != nil {
	// 	data := ws.NewCloseFrameBody(ws.StatusCode(disconnect.Code), disconnect.CloseText())
	// 	_ = wsutil.WriteServerMessage(t.conn, ws.OpClose, data)
	// 	return t.conn.Close()
	// }
	data := ws.NewCloseFrameBody(ws.StatusNormalClosure, "")
	_ = wsutil.WriteServerMessage(t.conn, ws.OpClose, data)
	return t.conn.Close()
}
