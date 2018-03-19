package MediatorService

import (
	"net"
	"strings"
)

type Client struct {
	Cconn  net.Conn
	IsDead chan bool
	Error  chan bool
	Recv   chan []byte
	Send   chan []byte
}

//这里接收Client的数据
func (self *Client) Read() {

	data := make([]byte, 10240)

	for {
		n, err := self.Cconn.Read(data)

		if err != nil {
			Log(self.Cconn.RemoteAddr().String(), " connection error: ", err)
			if strings.Contains(err.Error(), "EOF") {
				self.Error <- true
			}
			return
		}
		//因为读取Internet连接的状态，所以不能讲Read goroutine关闭
		//internet的链接，一般是短连接，假如Hacker
		self.Recv <- data[:n]
	}

}

//向Client发送数据
func (self *Client) Write() {

	data := make([]byte, 10240)

	for {

		select {
		case data = <-self.Send:
			//Log(self.Cconn.RemoteAddr().String(), "Send: \n", string(data))
			self.Cconn.Write(data)

		}

	}
}

//在另一个协程中监听端口函数
func CconnAccept(con net.Listener, Cconn chan net.Conn) {

	Log("Wait Client Connect ..")
	CorU, err := con.Accept()
	if err != nil {
		Log("Client Accept Error!")
		//out of goroutine
		return
	}
	Log("Client connect Success..")
	Cconn <- CorU

}
