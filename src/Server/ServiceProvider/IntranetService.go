package ServiceProvider

import (
	"log"
	"net"
	"strings"
)

type Service struct {
	Cconn  net.Conn
	IsDead chan bool
	Error  chan bool
	Recv   chan []byte
	Send   chan []byte
}

func Log(v ...interface{}) {
	log.Println(v...)
}

//这里Read接受Local Service的数据
func (self *Service) Read() {

	data := make([]byte, 2048)

	for {
		n, err := self.Sconn.Read(data)

		if err != nil {
			Log(self.Cconn.RemoteAddr().String(), " connection error: ", err)

			if strings.Contains(err.Error(), "EOF") {
				self.IsDead <- true
			}

			return
		}

		self.Recv <- data[:n]
	}

}

//向local service发送数据
func (self *Service) Write() {

	data := make([]byte, 10240)

	for {

		select {
		case data = <-self.Send:
			self.Cconn.Write(data)
		case <-self.IsDead:
			//Do Someting
			self.Error <- true
		}

	}
}
