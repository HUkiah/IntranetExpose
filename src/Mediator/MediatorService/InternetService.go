package MediatorService

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

type Client struct {
	Cconn    net.Conn
	Error    chan bool
	WritDead chan bool
	Recv     chan []byte
	Send     chan []byte
}

//这里Read接受Client的数据
func (self *Client) Read() {

	for {
		self.Cconn.SetReadDeadline(time.Now().Add(time.Second * 30))
		data := make([]byte, 10240)
		n, err := self.Cconn.Read(data)

		if err != nil {
			fmt.Println("Read Client Data display Error...")
			fmt.Println("Error:", err.Error())
			self.Error <- true
			break
		}

		self.Recv <- data[:n]
	}

}

//向Client发送数据
func (self *Client) Write() {

	data := make([]byte, 10240)

	for {
		select {

		case data = <-self.Send:
			self.Cconn.Write(data)
		case <-self.WritDead:
			fmt.Println("self.writDead...")
			goto EXIT
		}
	}

EXIT:
}

//在另一个协程中监听端口函数
func CconnAccept(con net.Listener, Cconn chan net.Conn) {
	fmt.Println("Wait Client Connect ..")
	CorU, err := con.Accept()
	if err != nil {
		fmt.Println("Client Accept Error!")
		//forced out of goroutine
		runtime.Goexit()
	}
	fmt.Println("Client connect Success!")
	Cconn <- CorU
}
