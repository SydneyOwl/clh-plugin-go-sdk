package clhplugin

import (
	"context"
	"encoding/binary"
	"errors"
	"github.com/Microsoft/go-winio"
	"io"
	"net"
	"runtime"

	"google.golang.org/protobuf/proto"
)

const maxDelimitedMessageSize = 16 << 20 // 16MB

func writeDelimitedMessage(w io.Writer, msg proto.Message) error {
	bin, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	var lenBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(bin)))
	if _, err = w.Write(lenBuf[:n]); err != nil {
		return err
	}
	_, err = w.Write(bin)
	return err
}

func readDelimitedMessage(r io.Reader, msg proto.Message) error {
	size, err := readUvarint(r)
	if err != nil {
		return err
	}
	if size > maxDelimitedMessageSize {
		return errors.New("delimited protobuf message too large")
	}

	buf := make([]byte, size)
	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}
	return proto.Unmarshal(buf, msg)
}

func readUvarint(r io.Reader) (uint64, error) {
	var (
		x uint64
		s uint
		b [1]byte
	)
	for i := 0; i < binary.MaxVarintLen64; i++ {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}
		if b[0] < 0x80 {
			if i == binary.MaxVarintLen64-1 && b[0] > 1 {
				return 0, errors.New("varint overflow")
			}
			return x | uint64(b[0])<<s, nil
		}
		x |= uint64(b[0]&0x7f) << s
		s += 7
	}
	return 0, errors.New("varint overflow")
}

func dialPipe(ctx context.Context, networkPath string) (net.Conn, error) {
	if runtime.GOOS == "windows" {
		return winio.DialPipeContext(ctx, networkPath)
	}
	var d net.Dialer
	return d.DialContext(ctx, "unix", networkPath)
}
