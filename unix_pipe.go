// unix_socket.go
//go:build !windows
// +build !windows

package clhplugin

import (
	"context"
	"net"
)

func dialPipe(ctx context.Context, networkPath string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, "unix", networkPath)
}
