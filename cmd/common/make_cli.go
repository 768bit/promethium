package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/websocket"

	client "github.com/768bit/promethium/api/client"
)

type OutboundJsonMessage struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	Payload   map[string]interface{} `json:"payload"`
}

type InboundJsonMessage struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	Payload   map[string]interface{} `json:"payload"`
	Code      int                    `json:"code"`
}

func MakeClient(host string, port int, path string) *client.Promethium {
	fullHost := fmt.Sprintf("%s:%d", host, port)
	transport := httptransport.New(fullHost, "", nil)
	return client.New(transport, strfmt.Default)
}

func MakeClientUnix() *client.Promethium {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/tmp/promethium.sock")
			},
		},
	}
	transport := httptransport.NewWithClient("unix", "", []string{"http"}, &httpc)
	transport.Transport = httpc.Transport
	//transport.SetDebug(true)
	return client.New(transport, strfmt.Default)
}

func MakeWebSocketClient(host string, port int, path string) (*websocket.Conn, error) {
	fullHost := fmt.Sprintf("%s:%d", host, port)
	u := url.URL{Scheme: "ws", Host: fullHost, Path: path}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func MakeWebSocketClientUnix(path string) (*websocket.Conn, error) {
	if websocket.DefaultDialer == nil {
		return nil, errors.New("The default dialler is nil!")
	}
	netDialer := &net.Dialer{}
	websocket.DefaultDialer.NetDial = func(network, addr string) (net.Conn, error) {
		println("Connecting to Unix Socket")
		return netDialer.DialContext(context.Background(), "unix", "/tmp/promethium.sock")
	}
	//nc, err := websocket.DefaultDialer.NetDial("unix", "/tmp/promethium")
	// nc, err := net.Dial("unix", "/tmp/promethium")
	// if err != nil {
	// 	return nil, err
	// }
	u := url.URL{Scheme: "ws", Host: "unix", Path: path}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	//c, _, err := websocket.NewClient(nc, &u, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	return c, nil
}
