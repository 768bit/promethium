package vmm

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/768bit/promethium/cmd/common"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"
)

var NullByte byte = 0
var STDINFILE = os.Stdin
var STDINFILENO = 0

func makePayload(inArr []byte) []byte {
	return append(inArr, NullByte)
}

var (
	writeWaitDuration time.Duration = time.Duration(400 * time.Millisecond)
	readWaitDuration  time.Duration = time.Duration(400 * time.Millisecond)
)

func setSaneTermMode() {
	raw, err := unix.IoctlGetTermios(STDINFILENO, unix.TCGETS)
	if err != nil {
		println(err.Error())
		return
	}
	rawState := *raw
	rawState.Iflag &^= unix.IGNBRK | unix.INLCR | unix.IGNCR | unix.IUTF8 | unix.IXOFF | unix.IUCLC | unix.IXANY
	rawState.Iflag |= unix.BRKINT | unix.ICRNL | unix.IMAXBEL
	rawState.Oflag |= unix.OPOST | unix.ONLCR
	rawState.Oflag &^= unix.OLCUC | unix.OCRNL | unix.ONOCR | unix.ONLRET
	rawState.Cflag |= unix.CREAD
	err = unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, &rawState)

	if err != nil {
		println(err.Error())
	}

	exec.Command("stty", "-F", "/dev/tty", "sane").Run()

}

var EOF_CHAR = byte(0x4)

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
		ws.SetCloseHandler(func(code int, text string) error {
			fmt.Printf("WebSocket Closed: %d : %s\n", code, text)
			return nil
		})
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

		raw, err := unix.IoctlGetTermios(STDINFILENO, unix.TCGETS)
		if err != nil {
			return err
		}
		rawState := *raw

		rawState.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
		// t.Iflag &^= BRKINT | ISTRIP | ICRNL | IXON // Stevens RAW
		rawState.Oflag &^= unix.OPOST
		rawState.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
		rawState.Cflag &^= unix.CSIZE | unix.PARENB
		rawState.Cflag |= unix.CS8
		rawState.Cc[unix.VMIN] = 1
		rawState.Cc[unix.VTIME] = 0

		err = unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, &rawState)

		if err != nil {
			return err
		}

		defer setSaneTermMode()

		// disable input buffering
		//exec.Command("stty", "-F", "/dev/tty", "cbreak", "isig", "min", "1").Run()
		// do not display entered characters on the screen
		//exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
		doExit := false
		go func() {

			inScanner := bufio.NewReaderSize(STDINFILE, 1)
			ibuff := make([]byte, 1)
			for {

				//ws.SetReadDeadline(time.Now().Add(writeWaitDuration))
				wo, err := ws.NextWriter(2)
				if err != nil {
					println("next_write:", err.Error())
					return
				}

				n, err := inScanner.Read(ibuff)
				if err != nil {
					println("stdin_read:", err.Error())
					return
				} else if n > 1 {
					println("Larger")
				}

				//check if the value is ^D

				if ibuff[0] == EOF_CHAR {
					print("\r\nExiting...\r\n")
					doExit = true
				}

				_, err = wo.Write(ibuff[:n])
				//wo.Write()

				//	_, err = io.Copy(wo, STDINFILE)

				// err = ws.WriteJSON(&common.OutboundJsonMessage{
				// 	ID:        "",
				// 	Operation: "console-input",
				// 	Payload: map[string]interface{}{
				// 		"input": string(b),
				// 	},
				// })
				if err != nil {
					wo.Close()
					println("stdin_write:", err.Error())
					return
				}
				wo.Close()
				if doExit {
					ws.Close()
					return
				}
			}
		}()
		buff := make([]byte, 1024)
		for {
			//ws.SetReadDeadline(time.Now().Add(readWaitDuration))
			mt, rd, err := ws.NextReader()

			//mt, msg, err := ws.ReadMessage()
			if err != nil {
				if doExit {
					return nil
				}
				return err
			}
			if mt == 2 {
				n, err := rd.Read(buff)
				if err != nil {
					if doExit {
						return nil
					}
					return err
				}
				// if err != nil {
				// 	println("read:", err.Error())
				// 	break
				// } else if n > 0 {
				// 	os.Stdout.Write(buff[:n])
				// }
				//_, err = io.Copy(os.Stdout, rd)
				_, err = os.Stdout.Write(buff[:n])
				if err != nil {
					if doExit {
						return nil
					}
					return err
				}
			} else {
				println("differ mt")
			}

			// inboundMsg := &common.InboundJsonMessage{}
			// err = ws.ReadJSON(inboundMsg)
			// if err != nil {
			// 	return err
			// }
			// switch inboundMsg.Operation {
			// case "console-output":
			// 	content := inboundMsg.Payload["output"].(string)
			// 	os.Stdout.WriteString(content)
			// }
			//fmt.Println(resp)
		}
		return nil

	},
}
