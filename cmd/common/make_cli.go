package common

import (
  "context"
  "fmt"
  "github.com/768bit/promethium/api/client"
  "net"
  "net/http"
)

func MakeClient(host string, port int, path string) *client.APIClient {
  cfg := client.NewConfiguration()
  cfg.Host = fmt.Sprintf("%s:%d", host, port)
  cfg.BasePath = fmt.Sprintf("%s:%d%s", host, port, path)
  return client.NewAPIClient(cfg)
}

func MakeClientUnix() *client.APIClient {
  httpc := http.Client{
    Transport: &http.Transport{
      DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
        return net.Dial("unix", "/tmp/promethium")
      },
    },
  }
  cfg := client.NewConfiguration()
  cfg.HTTPClient = &httpc
  return client.NewAPIClient(cfg)
}
