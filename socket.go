package ljus

import (
	"errors"
	"net"
	"os"
)

// SocketListener is a [net.Listener] for unix sockets.
//
// See [ListenSocket].
type SocketListener struct {
	net.Listener
	Path string
}

// ListenSocket returns a new [SocketListener].
func ListenSocket(path string, permissions os.FileMode) (*SocketListener, error) {
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	err = os.Chmod(path, permissions)
	if err != nil {
		errl := l.Close()
		if errl != nil {
			err = errors.Join(err, errl)
		}
		return nil, err
	}
	return &SocketListener{l, path}, nil
}

// Close the [SocketListener].
func (s *SocketListener) Close() error {
	err := s.Listener.Close()
	errRm := os.Remove(s.Path)
	if err != nil {
		if errRm != nil {
			err = errors.Join(err, errRm)
		}
		return err
	}
	if errRm != nil && !os.IsNotExist(errRm) {
		return errRm
	}
	return nil
}
