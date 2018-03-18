package MediatorService

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

type Mediator struct {
	Mconn    net.Conn
	Heart    chan bool
	Error    chan bool
	WritDead chan bool
	Recv     chan []byte
	Send     chan []byte
}

//because mediator attribute ,so Read data is not processed. write same
//这里read接受serviceProvider的数据
func (self *Mediator) Read() {

	var data []byte = make([]byte, 10240)

TOP:
	try := 0

	self.Mconn.SetReadDeadline(time.Now().Add(time.Second * 30))

RETRY:
	n, err := self.Mconn.Read(data)

	//已经Temporary Error处理 timeout错误
	if err != nil {
		e, ok := err.(net.Error)
		if !ok || !e.Temporary() || try >= 3 {
			//尝试3次后，标记为错误
			self.Error <- true
			self.Heart <- true
			goto EXIT
		}
		try++
		self.Mconn.SetReadDeadline(time.Now().Add(0 * time.Second))
		goto RETRY
	}

	//收到心跳包
	if data[0] == 'h' && data[1] == 'h' {

		self.Mconn.Write([]byte("hh"))
		goto TOP
	}

	self.Recv <- data[:n]

	//每次接收到数据，重置timeout计时
	goto TOP

EXIT:
}

func (self *Mediator) Write() {

	var data []byte = make([]byte, 10240)

	for {

		select {

		case data = <-self.Send:
			self.Mconn.Write(data)
		case <-self.WritDead:
			//Service shutdown Connect
			goto EXIT

		}

	}

EXIT:
}

//mediator listen function
func MediatorAccept(conn net.Listener) net.Conn {

	fmt.Println("Wait ServiceProvider Connect ..")
	CorU, err := conn.Accept()
	if err != nil {
		fmt.Println("Mediator Accept Error!")
		//forced out of goroutine
		runtime.Goexit()
	}
	fmt.Println("ServiceProvider connect Success!")
	return CorU
}

// //长连接
// func handleConnection(conn net.Conn, timeout int) {
//
// 	buffer := make([]byte, 2048)
//
// 	for {
//
// 		n, err := conn.Read(buffer)
//
// 		if err != nil {
// 			//logErr(conn.RemoteAddr().String(), "Connection Error:", err)
// 			return
// 		}
//
// 		Data := (buffer[:n])
//
// 		msg := make(chan byte, 0)
// 		postData := make(chan byte, 0)
//
// 		//心跳检测
// 		go HeartBeating(conn, msg, timeout)
//
// 		//检测每次Client是否有数据传来
// 		go GravelChannel(Data, msg)
//
// 		//记录日志
//
// 	}
// }
//
// //根据数据监控，判断是否在设定的时间内发来信息
// func HeartBeating(conn net.Conn, msg chan byte, timeout int) {
//
// 	select {
// 	case <-msg:
// 		//记录日志
// 		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
// 		break
// 	case <-time.After(time.Second * 5):
// 		//记录日志
// 		conn.Close()
// 	}
// }
//
// //数据监控
// func GravelChannel(n []byte, msg chan byte) {
// 	for _, v := range n {
// 		msg <- v
// 	}
// 	close(msg)
// }
