package main

import (
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"uds/app"
	"uds/tcp"
	"uds/udss"
	"uds/utils"
)

func main() {
	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)

	logfile, _ := os.Create("main.log")
	c := make(chan os.Signal)
	channel_uds := make(chan int)
	channel_tcp := make(chan int)

	defer func() {
		close(c)
		close(channel_uds)
		close(channel_tcp)
	}()

	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				channel_uds <- utils.SIG_EXIT
				channel_tcp <- utils.SIG_EXIT
			case syscall.SIGUSR1:
				channel_uds <- utils.SIG_USER1
				channel_tcp <- utils.SIG_USER1
			case syscall.SIGUSR2:
				channel_uds <- utils.SIG_USER2
				channel_tcp <- utils.SIG_USER2
			}
		}
	}()

	udss.Initialize(app.NewApp(app.TYPE_UDS), chan<- int(channel_uds), (<-chan int)(channel_uds))
	tcp.Initialize(app.NewApp(app.TYPE_TCP), chan<- int(channel_tcp), (<-chan int)(channel_tcp))

	go udss.Start()
	go tcp.Start("192.168.1.12", "2111")

	tc := false
	for {
		select {
		case ret := <-channel_uds:
			io.WriteString(logfile, "uds exit with:"+strconv.Itoa(ret)+"\n")
			if tc {
				os.Exit(0)
			} else {
				tc = true
				channel_tcp <- utils.SIG_EXIT
			}
		case ret := <-channel_tcp:
			io.WriteString(logfile, "tcp exit with:"+strconv.Itoa(ret)+"\n")
			if tc {
				os.Exit(0)
			} else {
				tc = true
				channel_uds <- utils.SIG_EXIT
			}
		}
	}
}
