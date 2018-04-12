package udss

import (
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/ifig4fireghost/uds/utils"
)

const (
	UDS_NORMAL_EXIT = iota
	UDS_CREATE_ERROR
	UDS_ACCEPT_ERROR
	UDS_READ_ERROR
	UDS_WRITE_ERROR
)

const (
	SS_NOT_INIT = -1
)

const (
	TAG = "UDS"
)

type App interface {
	Do(conn net.Conn)
	OnSignal(sig int)
}

var logfile *os.File
var mapp App
var channel_to_main chan<- int
var channel_from_main <-chan int

var once sync.Once

func fatal(err error, ret int) {
	io.WriteString(logfile, err.Error()+"\n")
	channel_to_main <- ret
}

func ReceivedSignal(sig int) {
	mapp.OnSignal(sig)
	io.WriteString(logfile, TAG+":received signal:"+strconv.Itoa(sig)+"\n")
	if sig == utils.SIG_EXIT {
		logfile.Sync()
		logfile.Close()
		channel_to_main <- UDS_NORMAL_EXIT
	}
}

func setup(app App, ch chan<- int, cf <-chan int) {
	logfile, _ = os.Create("uds.log")
	mapp = app
	channel_to_main = ch
	channel_from_main = cf
}

func Initialize(app App, ch chan<- int, cf <-chan int) {
	once.Do(func() { setup(app, ch, cf) })
}

func Start() {
	if mapp == nil {
		io.WriteString(logfile, TAG+":nil receiver!\n")
		channel_to_main <- SS_NOT_INIT
	}
	go func() {
		for {
			select {
			case sig := <-channel_from_main:
				ReceivedSignal(sig)
			}
		}
	}()

	if logfile == nil {
		fatal(errors.New("not initilaized first."), SS_NOT_INIT)
	}
	path := utils.Generate("ifig-graduate-project")
	syscall.Unlink(path)
	socket, err := net.Listen("unix", path)
	if err != nil {
		fatal(err, UDS_CREATE_ERROR)
	}
	defer syscall.Unlink(path)
	io.WriteString(logfile, "UDSServer runnning...\n")
	for {
		client, err := socket.Accept()
		if err != nil {
			fatal(err, UDS_ACCEPT_ERROR)
		}
		mapp.Do(client)
	}
}
