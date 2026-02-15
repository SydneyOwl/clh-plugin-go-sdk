// windows_pipe.go
//go:build windows

package pluginsdk

import (
	"github.com/Microsoft/go-winio"
	"net"
)

const PipePath = "\\\\.\\pipe\\clh.plugin"

func dialPipe() (net.Conn, error) {
	return winio.DialPipe(PipePath, nil)
}
