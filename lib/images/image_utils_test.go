package images

import (
  "github.com/768bit/promethium/lib/cloudconfig"
  "testing"
)

func TestMakeCloudInitImage(t *testing.T) {
  _, err := MakeCloudInitImage("testing", &cloudconfig.MetaDataNetworkConfig{}, &cloudconfig.UserData{})
  if err != nil {
    t.Error(err)
  }
}
