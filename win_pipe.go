// windows_pipe.go
//go:build windows

package clhplugin

import (
	"context"
	"net"

	"github.com/Microsoft/go-winio"
)

func dialPipe(ctx context.Context, networkPath string) (net.Conn, error) {
	return winio.DialPipeContext(ctx, networkPath)
}
