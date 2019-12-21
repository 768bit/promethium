package vmm

import (
	"bufio"
	"os"

	"github.com/768bit/promethium/cmd/common"
	"github.com/urfave/cli/v2"
)

var InstanceConsoleCommand = cli.Command{
	Name:  "console",
	Usage: "Get instance console.",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {
		id := c.Args().Get(0)
		ws, err := common.MakeWebSocketClientUnix("/consolews")
		if err != nil {
			return err
		}
		defer ws.Close()
		err = ws.WriteJSON(common.OutboundJsonMessage{
			ID:        "",
			Operation: "connect-console",
			Payload: map[string]interface{}{
				"id": id,
			},
		})
		if err != nil {
			return err
		}
		go func() {
			inScanner := bufio.NewReader(os.Stdin)
			obuff := make([]byte, 1024)
			for {

				n, err := inScanner.Read(obuff)
				if err != nil {
					println("read:", err.Error())
					return
				} else if n == 0 {
					continue
				}

				err = ws.WriteJSON(&common.OutboundJsonMessage{
					ID:        "",
					Operation: "console-input",
					Payload: map[string]interface{}{
						"input": string(obuff[:n]),
					},
				})
				if err != nil {
					println("write:", err.Error())
					return
				}
			}
		}()
		for {
			inboundMsg := &common.InboundJsonMessage{}
			err = ws.ReadJSON(inboundMsg)
			if err != nil {
				return err
			}
			switch inboundMsg.Operation {
			case "console-output":
				content := inboundMsg.Payload["output"].(string)
				os.Stdout.WriteString(content)
			}
			//fmt.Println(resp)
		}
		return nil

	},
}
