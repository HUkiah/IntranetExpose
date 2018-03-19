package ServiceProvider

import (
	"net"
	"strings"
	"time"
)

type Mediator struct {
	Mconn  net.Conn
	IsDead chan bool
	Error  chan bool
	Recv   chan []byte
	Send   chan []byte
}

//这里Read接受IntranetService的数据
func (self *Mediator) Read() {

	var data []byte = make([]byte, 2048)

TOP:
	try := 0

	self.Mconn.SetReadDeadline(time.Now().Add(time.Second * 300))

	for {
		n, err := self.Mconn.Read(data)

		//记录是否断开
		if err != nil {
			Log(self.Mconn.RemoteAddr().String(), " connection error: ", err)

			//不是一个网络错误，临时错误，已经试了3次
			if strings.Contains(err.Error(), "timeout") && try <= 5 {
				//取消超时
				self.Mconn.SetReadDeadline(time.Time{})
				//如果因为网络问题，延长等待30s。没有导致没有收到，则发送3次心跳
				if try >= 3 {
					//try 3 次后，发送3心跳包
					try++
					self.Send <- []byte("xx")
					self.Mconn.SetReadDeadline(time.Now().Add(20 * time.Second))
					continue
				}
				try++
				self.Mconn.SetReadDeadline(time.Now().Add(20 * time.Second))
				continue
			}

			self.IsDead <- true
			return
		}

		//Log(self.Mconn.RemoteAddr().String(), "Receive:\n", string(data[:n]))

		//心跳包，原样返回
		if data[0] == 'x' && data[1] == 'x' {

			goto TOP
		} else if data[0] == 'h' && data[1] == 'h' {

			self.Send <- data[:n]
			goto TOP
		} else {
			self.Recv <- data[:n]
		}

		//接收到数据，重置ReadDeadLine
		goto TOP
	}

}

//
func (self *Mediator) Write() {

	data := make([]byte, 10240)

	for {

		select {
		case data = <-self.Send:
			//Log(self.Mconn.RemoteAddr().String(), "Send:\n", string(data))
			self.Mconn.Write(data)
		case <-self.IsDead:
			//Do Something
			self.Error <- true
			//return
		}

	}
}
