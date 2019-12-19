package storage

import (
  "testing"
)

func TestProcessCreate(t *testing.T) {
  zfs, err := NewZfsStorageDrive("testing", 1024)

  if err != nil {
    t.Errorf("Error making/loading zfs storage: %s", err.Error())
    return
  }
  err = zfs.Destroy()
  if err != nil {
    t.Errorf("Error destorying zfs %s", err.Error())
  }
}
