// unix_socket.go
//go:build !windows
// +build !windows

package pluginsdk

import "net"

const PipePath = "/tmp/clh.plugin"

func dialPipe() (net.Conn, error) {
	return net.Dial("unix", PipePath)
}
