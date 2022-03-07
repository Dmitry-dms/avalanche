package websocket

import (
	"io/ioutil"
	"net"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pkg/errors"
)

type GobwasWebsocket struct {
	closed  bool
	conn    net.Conn
}

func NewWebsocketTransport(conn net.Conn) GobwasWebsocket {
	return GobwasWebsocket{
		conn:    conn,
	}
}
func (t *GobwasWebsocket) IsClosed() bool {
	return t.closed
}

func (t GobwasWebsocket) Read() ([]byte, bool, error) {
	h, r, err := wsutil.NextReader(t.conn, ws.StateServerSide)
	if err != nil {
		return nil, true, errors.Wrap(err, "reader error")
	}

	if h.OpCode == ws.OpPing {
		return nil, false, wsutil.WriteServerMessage(t.conn, ws.OpPong, []byte{})
	}
	if h.OpCode == ws.OpPong {
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

func (t GobwasWebsocket) Write(msg []byte) error {
	if err := wsutil.WriteServerMessage(t.conn, ws.OpText, msg); err != nil {
		return err
	}
	return nil
}

func (t GobwasWebsocket) Close() error {
	// if t.closed {
	// 	return nil
	// }
	// t.closed = true
	//close(t.closeCh)
	data := ws.NewCloseFrameBody(ws.StatusNormalClosure, "closing connection")
	_ = wsutil.WriteServerMessage(t.conn, ws.OpClose, data)
	return t.conn.Close()
}
