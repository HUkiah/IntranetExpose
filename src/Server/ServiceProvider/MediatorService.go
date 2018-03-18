package ServiceProvider

import (
	"fmt"
	"net"
	"time"
)

type Mediator struct {
	Mconn    net.Conn
	Error    chan bool
	WritDead chan bool
	Recv     chan []byte
	Send     chan []byte
}

//这里Read接受IntranetService的数据
func (self *Mediator) Read() {

	var data []byte = make([]byte, 2048)

TOP:
	try := 0

	self.Mconn.SetReadDeadline(time.Now().Add(time.Second * 20))

RETRY:
	n, err := self.Mconn.Read(data)

	//已经Temporary Error处理 timeout错误
	if err != nil {
		e, ok := err.(net.Error)
		if !ok || !e.Temporary() || try >= 3 {

			if try <= 6 {
				//try 3 次后，发送3心跳包
				try++
				self.Mconn.Write([]byte("hh"))

				self.Mconn.SetReadDeadline(time.Now().Add(20 * time.Second))
				goto RETRY
			}
			//没有收到回应，应该是代理挂掉了
			fmt.Println("Read MediatorService have Error...")
			self.Error <- true
			goto EXIT
		}
		try++
		self.Mconn.SetReadDeadline(time.Now().Add(20 * time.Second))
		goto RETRY
	}

	//收到心跳包，reset
	if data[0] == 'h' && data[1] == 'h' {
		fmt.Println("package Heart..")
		goto TOP
	}

	self.Recv <- data[:n]

	//每次接收到数据，重置timeout计时
	goto TOP

EXIT:
}

//
func (self *Mediator) Write() {

	var data []byte = make([]byte, 2048)

	for {

		select {

		case data = <-self.Send:
			self.Mconn.Write(data)
			fmt.Println("Mediator->Remote Mediator..")
		case <-self.WritDead:
			fmt.Println("self.WritDead..")
			//Service shutdown Connect
			goto EXIT

		}

	}

EXIT:
}
