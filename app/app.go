package app

import (
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ifig4fireghost/uds/utils"
)

const (
	TYPE_UDS = iota
	TYPE_TCP
)

var logfile *os.File

func init() {
	logfile, _ = os.Create(".app-log")
}

type Processor interface {
	Do(conn net.Conn)
	OnSignal(sig int)
}

type UDSProcessor struct {
	processor *TCPProcessor
}

type TCPProcessor struct {
	is_connected bool

	conn *net.TCPConn
}

func NewApp(t int) Processor {
	switch t {
	case TYPE_UDS:
		return &UDSProcessor{nil}
	case TYPE_TCP:
		return &TCPProcessor{false, nil}
	default:
		return nil
	}
}

func (app *UDSProcessor) Do(conn net.Conn) {
	defer conn.Close()
	app.processor = NewApp(TYPE_TCP).(*TCPProcessor)
	io.WriteString(logfile, "Connection @"+time.Now().String()+"\n")
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
		data, err := utils.Read(recv)
		if err != nil {
			continue
		}

		io.WriteString(logfile, string(data)+"\n")
		cmds := strings.Split(string(data), "&")
		for _, v := range cmds {
			cmd := strings.Split(string(v), "=")
			switch cmd[0] {
			case "C":
				switch cmd[1] {
				case "KP":
					io.WriteString(logfile, "recv keep from "+conn.RemoteAddr().String()+"reset timer @"+time.Now().String()+"\n")
				case "QT":
					if app.processor.is_connected {
						app.processor.conn.Close()
						io.WriteString(logfile, "tcp connection closed\n")
					}
					goto OVER
				case "CT":
					if len(cmd) == 3 {
						if app.processor.is_connected {
							conn.Write(utils.Write("CA"))
						} else {
							if app.processor.Start(cmd[2]) {
								conn.Write(utils.Write("CD"))
							} else {
								io.WriteString(logfile, "connect failed\n")
								conn.Write(utils.Write("CF"))
							}
						}
					} else {
						conn.Write(utils.Write("PF"))
						io.WriteString(logfile, "connect: parameter num not match\n")
					}
				default:
					if len(cmd) == 2 {
						if !app.processor.is_connected {
							conn.Write(utils.Write("NC"))
							io.WriteString(logfile, "make connect first\n")
						} else {
							ret := app.processor.Get(cmd[1])
							if ret != nil {
								conn.Write(utils.Write(string(ret)))
							} else {
								conn.Write(utils.Write("NA"))
							}
						}
					}
				}
			}
		}
	}
OVER:
	io.WriteString(logfile, "connection closed\n")
}

func (app *UDSProcessor) OnSignal(sig int) {
	io.WriteString(logfile, "uds received signal:"+strconv.Itoa(sig)+"\n")
	app.processor.OnSignal(sig)
	if sig == utils.SIG_EXIT {
		logfile.Sync()
		logfile.Close()
	}
}

func (app *TCPProcessor) Do(conn net.Conn) {
	return
}

func read_data(conn *net.TCPConn) []byte {
	BufLength := 10240
	recv := make([]byte, 0)
	buf := make([]byte, BufLength)

	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return nil
	}
	if n > 0 {
		recv = append(recv, buf[:n]...)
	}
	return recv
}

func (app *TCPProcessor) Get(para string) []byte {
	_, err := app.conn.Write([]byte(para))
	if err != nil {
		return nil
	}
	return read_data(app.conn)
}

func (app *TCPProcessor) OnSignal(sig int) {
	io.WriteString(logfile, "tcp received signal:"+strconv.Itoa(sig)+"\n")
	if sig == utils.SIG_EXIT {
		logfile.Sync()
		logfile.Close()
	}
}

func (app *TCPProcessor) Start(server string) bool {
	conn, err := net.DialTimeout("tcp4", server, time.Second*3)
	if err != nil {
		return false
	} else {
		io.WriteString(logfile, "connect success\n")
		app.is_connected = true
		app.conn = conn.(*net.TCPConn)
		return true
	}
}
