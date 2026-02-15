// unix_socket.go
//go:build !windows
// +build !windows

package pluginsdk

import "net"

const PipePath = "/tmp/clh.plugin"

func dialPipe() (net.Conn, error) {
	conn, err := net.Dial("unix", PipePath)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
