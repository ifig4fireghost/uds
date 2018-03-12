package tcp

import (
	"fmt"
	"testing"

	"Radar/app"
)

func TestClient(t *testing.T) {
	channel_tcp := make(chan int)
	defer close(channel_tcp)

	Initialize(app.NewApp(app.TYPE_TCP), chan<- int(channel_tcp), (<-chan int)(channel_tcp))
	go Start("192.168.1.20", "2112")

	select {
	case ret := <-channel_tcp:
		fmt.Println("tcp exit with:", ret)
	}
}
