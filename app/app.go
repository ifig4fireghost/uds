package app

import (
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ifig4fireghost/uds/tcp"
	"github.com/ifig4fireghost/uds/utils"
)

const (
	TYPE_UDS = iota
	TYPE_TCP
)

var logfile *os.File

func init() {
	logfile, _ = os.Create("app.log")
}

type Processor interface {
	Do(conn net.Conn)
	OnSignal(sig int)
}

type UDSProcessor struct {
}

type TCPProcessor struct {
}

func NewApp(t int) Processor {
	switch t {
	case TYPE_UDS:
		return &UDSProcessor{}
	case TYPE_TCP:
		return &TCPProcessor{}
	default:
		return nil
	}
}

func (app *UDSProcessor) Do(conn net.Conn) {
	defer conn.Close()

	io.WriteString(logfile, "Eastablished connection @"+time.Now().String()+"\n")
	BufLength := 1024
	for {
		recv := make([]byte, 0)
		buf := make([]byte, BufLength)
	RECV_LOOP:
		for {
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				io.WriteString(logfile, "timeout\n")
				goto OVER
			}
			if n > 0 {
				recv = append(recv, buf[:n]...)
				if n != BufLength {
					break RECV_LOOP
				}
			}
		}
		io.WriteString(logfile, string(recv)+"\n")
		data, err := utils.Decode(string(recv))
		if err != nil {
			continue
		}

		io.WriteString(logfile, string(data)+"\n")
		cmds := strings.Split(string(data), "&")
		for _, v := range cmds {
			cmd := strings.Split(string(v), "=")
			switch cmd[0] {
			case "cmd":
				switch cmd[1] {
				case "keep":
					io.WriteString(logfile, "recv keep from "+conn.RemoteAddr().String()+"reset timer @"+time.Now().String()+"\n")
				case "quit":
					goto OVER
				case "connect":
					if len(cmd) == 3 {
						go tcp.Start(cmd[2], cmd[3])
					} else {
						io.WriteString(logfile, "connect: parameter num must be 3\n")
					}
				default:
					conn.Write([]byte(utils.Encode([]byte("NotSupport:" + cmd[1]))))
				}
			case "data":
				conn.Write([]byte(utils.Encode([]byte(cmd[1]))))
			}
		}
	}
OVER:
	io.WriteString(logfile, "connection closed\n")
}

func (app *UDSProcessor) OnSignal(sig int) {
	io.WriteString(logfile, "uds received signal:"+strconv.Itoa(sig)+"\n")
	if sig == utils.SIG_EXIT {
		logfile.Sync()
		logfile.Close()
	}
}

func (app *TCPProcessor) Do(conn net.Conn) {
	defer func() {
		conn.Close()
		io.WriteString(logfile, "connection closed\n")
	}()
	io.WriteString(logfile, "Eastablished connection with "+conn.RemoteAddr().String()+"@"+time.Now().String()+"\n")
}

func (app *TCPProcessor) OnSignal(sig int) {
	io.WriteString(logfile, "tcp received signal:"+strconv.Itoa(sig)+"\n")
	if sig == utils.SIG_EXIT {
		logfile.Sync()
		logfile.Close()
	}
}
