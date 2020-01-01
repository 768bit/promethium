package common

import "time"

import "io"

type VmmProcess interface {
	GetStatus() string
	Wait() error
	Console() (io.ReadCloser, io.ReadCloser, io.WriteCloser, error)
	Start() error
	Stop() error
	Shutdown() error
	ShutdownTimeout(timeout time.Duration) error
	Restart() error
	Reset() error
}
