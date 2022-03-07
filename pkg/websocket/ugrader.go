package websocket

import (
	"fmt"
	"io"

	"github.com/gobwas/ws"
)

type GobwasUpgrader struct {
	//Upgrader ws.Upgrader
}

// func NewGobwasUpgrader(readBufferSize, writeBufferSize int, userId, companyName *string, parseFunc func(accessToken string) (string, error)) GobwasUpgrader {
// 	u := ws.Upgrader{
// 		ReadBufferSize:  readBufferSize,
// 		WriteBufferSize: writeBufferSize,
// 		OnHeader: func(key, value []byte) error {
// 			if string(key) == "User" {
// 				*userId = string(value)
// 			} else if string(key) == "Token" {
// 				var err error
// 				companyN, err := parseFunc(string(value))
// 				if err != nil {
// 					return ws.RejectConnectionError(
// 						ws.RejectionReason(fmt.Sprintf("bad token: %s", err)),
// 						ws.RejectionStatus(400))
// 				}
// 				*companyName = companyN
// 			}
// 			return nil
// 		},
// 	}
// 	return GobwasUpgrader{Upgrader: u}
// }

func (g *GobwasUpgrader) Upgrade(conn io.ReadWriter, readBufferSize, writeBufferSize int, userId, companyName *string, parseFunc func(accessToken string) (string, error)) error {
	u := ws.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		OnHeader: func(key, value []byte) error {
			if string(key) == "User" {
				*userId = string(value)
			} else if string(key) == "Token" {
				var err error
				companyN, err := parseFunc(string(value))
				if err != nil {
					return ws.RejectConnectionError(
						ws.RejectionReason(fmt.Sprintf("bad token: %s", err)),
						ws.RejectionStatus(400))
				}
				*companyName = companyN
			}
			return nil
		},
	}
	_, err := u.Upgrade(conn)
	return err
}
