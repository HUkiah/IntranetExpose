package ServiceProvider

import (
	"fmt"
	"net"
)

type Service struct {
	Sconn    net.Conn
	Error    chan bool
	WritDead chan bool
	Recv     chan []byte
	Send     chan []byte
}

//这里Read接受Local Service的数据
func (self *Service) Read() {

	data := make([]byte, 2048)

	for {
		n, err := self.Sconn.Read(data)

		if err != nil {
			fmt.Println("Read Local Service have Error...")
			self.Error <- true
			break
		}

		self.Recv <- data[:n]
	}

}

//向local service发送数据
func (self *Service) Write() {

	data := make([]byte, 2048)

	for {
		select {

		case data = <-self.Send:
			self.Sconn.Write(data)
		case <-self.WritDead:
			goto EXIT
		}
	}

EXIT:
}
