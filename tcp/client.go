package tcp

import (
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ifig4fireghost/uds/utils"
)

type App interface {
	Do(conn net.Conn)
	OnSignal(sig int)
}

const (
	TCP_NORMAL_EXIT = iota
	TCP_CREATE_ERROR
	TCP_READ_ERROR
	TCP_WRITE_ERROR
)

const (
	TAG = "tcp client"
)

var logfile *os.File
var once sync.Once
var mapp App
var channel_to_main chan<- int
var channel_from_main <-chan int

func fatal(err error, ret int) {
	io.WriteString(logfile, "Fatal error: "+err.Error()+"\n")
	channel_to_main <- ret
}

func ReceivedSignal(sig int) {
	mapp.OnSignal(sig)
	io.WriteString(logfile, TAG+":received signal:"+strconv.Itoa(sig)+"\n")
	if sig == utils.SIG_EXIT {
		logfile.Sync()
		logfile.Close()
		channel_to_main <- TCP_NORMAL_EXIT
	}
}

func setup(app App, ch chan<- int, cf <-chan int) {
	logfile, _ = os.Create("data.log")
	mapp = app
	channel_to_main = ch
	channel_from_main = cf
}

func Initialize(app App, ch chan<- int, cf <-chan int) {
	once.Do(func() { setup(app, ch, cf) })
}

func Start(ip string, port string) {
	go func() {
		for {
			select {
			case sig := <-channel_from_main:
				ReceivedSignal(sig)
			}
		}
	}()

	server := ip + ":" + port
	conn, err := net.DialTimeout("tcp4", server, time.Second*3)
	if err != nil {
		fatal(err, TCP_CREATE_ERROR)
	} else {
		io.WriteString(logfile, "connect success\n")
		mapp.Do(conn)
	}
}

//func heartbeat(conn *net.TCPConn) {
//	timer := time.NewTimer(5 * time.Second)
//	<-timer.C
//	_, err := conn.Write([]byte("kp"))
//	if err != nil {
//		io.WriteString(logfile, "heartbeat: "+err.Error()+"\n")
//		return
//	}
//	go heartbeat(conn)
//}

//func read(conn *net.TCPConn) []byte {
//	BufLength := 1024
//	data := make([]byte, 0)
//	buf := make([]byte, BufLength)
//	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
//	for {
//		n, err := conn.Read(buf)
//		if err != nil && err != io.EOF {
//			io.WriteString(logfile, "timeout\n")
//			return nil
//		}
//		if n > 0 {
//			data = append(data, buf[:n]...)
//			if n != BufLength {
//				break
//			}
//		}
//	}
//	return data
//}

//func handle(conn net.Conn) {
//	io.WriteString(logfile, conn.RemoteAddr().String()+"\n")
//	utils.Quit(TCP_NORMAL_EXIT)
//}
