package websocket

import (
	//"fmt"
	"context"

	"fmt"

	"io/ioutil"
	"net"
	"sync"

	//"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
)

type CustomWebsocketTransport struct {
	mu      sync.RWMutex
	closed  bool
	closeCh chan struct{}
	conn    net.Conn
}
type CustomCancelContext struct {
	context.Context
	ch <-chan struct{}
}

func NewCustomContext() CustomCancelContext {
	return CustomCancelContext{
		context.TODO(),
		make(<-chan struct{}),
	}
}
func (c CustomCancelContext) Done() <-chan struct{} {
	return c.ch
}

func (c CustomCancelContext) Err() error {
	select {
	case <-c.ch:
		return context.Canceled
	default:
		return nil
	}
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

func (t *CustomWebsocketTransport) Read() ([]byte, bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	h, r, err := wsutil.NextReader(t.conn, ws.StateServerSide)
	if err != nil {
		return nil, true, errors.Wrap(err, "reader error")
	}

	//TODO: make right header parsing between gorilla and gobwas (opCode is missing in gorilla)
	if h.OpCode == ws.OpPing {
		fmt.Println("called ping, sending pong")
		return nil, false, wsutil.WriteServerMessage(t.conn, ws.OpPong, []byte{})
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

func (t *CustomWebsocketTransport) Write(msg []byte) error {
	messageType := ws.OpText
	if err := wsutil.WriteServerMessage(t.conn, messageType, msg); err != nil {
		return err
	}
	return nil
	//for {
	// select {
	// case <-t.closeCh:
	// 		return nil
	// case msg := <-ch:
	// 	messageType := ws.OpText
	// 	if err := wsutil.WriteServerMessage(t.conn, messageType, []byte(msg)); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }
	//}
}

func (t *CustomWebsocketTransport) Close() error {
	//t.mu.Lock()
	//log.Println("CLOSING")
	if t.closed {
		//t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.closeCh <- struct{}{}
	close(t.closeCh)
	//t.mu.Unlock()
	// data := ws.NewCloseFrameBody(ws.StatusNormalClosure, "closing connection")
	// _ = wsutil.WriteServerMessage(t.conn, ws.OpClose, data)
	return t.conn.Close()
}
