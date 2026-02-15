// windows_pipe.go
//go:build windows

package pluginsdk

import (
	"github.com/Microsoft/go-winio"
	"log"
	"net"
)

const PipePath = "\\\\.\\pipe\\clh.plugin"

func dialPipe() (net.Conn, error) {
	conn, err := winio.DialPipe(PipePath, nil)
	if err != nil {
		log.Fatalf("Failed to connect to pipe: %v", err)
	}
	return conn, nil
}
