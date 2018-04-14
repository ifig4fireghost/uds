package main

import (
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/ifig4fireghost/uds/app"
	"github.com/ifig4fireghost/uds/udss"
	"github.com/ifig4fireghost/uds/utils"
)

func main() {
	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)

	logfile, _ := os.Create(".main-log")
	c := make(chan os.Signal)
	channel_uds := make(chan int)

	defer func() {
		close(c)
		close(channel_uds)
	}()

	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				channel_uds <- utils.SIG_EXIT
			}
		}
	}()

	udss.Initialize(app.NewApp(app.TYPE_UDS), chan<- int(channel_uds), (<-chan int)(channel_uds))

	go udss.Start()
	ret := <-channel_uds
	io.WriteString(logfile, "uds exit with:"+strconv.Itoa(ret)+"\n")
	os.Exit(0)
}
