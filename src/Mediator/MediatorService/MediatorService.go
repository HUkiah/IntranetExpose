package MediatorService

import (
	"log"
	"net"
	"runtime"
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

func Log(v ...interface{}) {
	log.Println(v...)
}

//because mediator attribute ,so Read data is not processed. write same
//这里read接受serviceProvider的数据
func (self *Mediator) Read() {

	var data []byte = make([]byte, 10240)

TOP:
	try := 0

	self.Mconn.SetReadDeadline(time.Now().Add(time.Second * 420))

	for {
		n, err := self.Mconn.Read(data)

		//已经Temporary Error处理 timeout错误
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
					self.Send <- []byte("hh")
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

		//Log(self.Mconn.RemoteAddr().String(), "Receive :\n", string(data[:n]))

		//心跳包，原样返回
		if data[0] == 'h' && data[1] == 'h' {

			goto TOP
		} else if data[0] == 'x' && data[1] == 'x' {

			self.Send <- data[:n]
			goto TOP
		} else {
			self.Recv <- data[:n]
		}

	}

}

func (self *Mediator) Write() {

	data := make([]byte, 10240)

	for {

		select {
		case data = <-self.Send:
			//Log(self.Mconn.RemoteAddr().String(), "Send: \n", string(data))
			self.Mconn.Write(data)
		case <-self.IsDead:
			//Do Something
			self.Error <- true

		}

	}
}

//mediator listen function
func MediatorAccept(conn net.Listener) net.Conn {

	Log("Wait ServiceProvider Connect ..")
	CorU, err := conn.Accept()
	if err != nil {
		Log("Mediator Accept Error!")
		//forced out of goroutine
		runtime.Goexit()
	}
	Log("ServiceProvider connect Success!")
	return CorU
}
