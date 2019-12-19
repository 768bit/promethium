package interfaces

import "time"

type VmmProcess interface {
  GetStatus() string
  Wait() error
  Start() error
  Stop() error
  Shutdown() error
  ShutdownTimeout(timeout time.Duration) error
  Restart() error
  Reset() error
}
