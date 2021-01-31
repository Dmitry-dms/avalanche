package websocket

import (
	"fmt"
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
	//t.mu.Lock()
	//defer t.mu.Unlock()
	fmt.Println("reading msg")
	// h, r, err := wsutil.NextReader(t.conn, ws.StateServerSide)
	// if err != nil {
	// 	t.Close()
	// 	fmt.Println("connection was closed")
	// 	return nil, false, err
	// }
	msg, err := wsutil.ReadClientBinary(t.conn)
	// if h.OpCode.IsControl() {
	// 	fmt.Println("connection is control")
	// 	return nil, true, wsutil.ControlFrameHandler(t.conn, ws.StateServerSide)(h, r)
	// }

	//data, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println("connection read binary")
		return nil, false, t.Close()
	}

	return msg, false, nil
}

func (t *CustomWebsocketTransport) Write(data []byte) error {
	select {
	case <-t.closeCh:
		return nil
	default:
		messageType := ws.OpBinary
		fmt.Println("writing msg")
		if err := wsutil.WriteServerMessage(t.conn, messageType, data); err != nil {
			//t.Close()
			return t.Close()
		}
		return nil
	}
}

// Close implementation.
func (t *CustomWebsocketTransport) Close() error {
	//fmt.Println("lock")
	//t.mu.Lock()
	//fmt.Println("lock 2")
	// if t.closed {
	// 	fmt.Println("closed")
	// 	//t.mu.Unlock()
	// 	return nil
	// }
	t.closed = true
	close(t.closeCh)
	//t.mu.Unlock()
	fmt.Println("unlock")
	// if disconnect != nil {
	// 	data := ws.NewCloseFrameBody(ws.StatusCode(disconnect.Code), disconnect.CloseText())
	// 	_ = wsutil.WriteServerMessage(t.conn, ws.OpClose, data)
	// 	return t.conn.Close()
	// }
	fmt.Println("closing")
	data := ws.NewCloseFrameBody(ws.StatusNormalClosure, "")
	_ = wsutil.WriteServerMessage(t.conn, ws.OpClose, data)
	return t.conn.Close()
}
