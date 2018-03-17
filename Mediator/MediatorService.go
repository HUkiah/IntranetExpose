package Mediator

import (

	"net"
	"runtime"
	"time"
)

type Mediator struct {
	Mconn    net.Conn
	Error    chan bool
	heart 	 chan bool
	writ     chan bool
	recv     chan []byte
	send     chan []byte
}

//because mediator attribute ,so Read data is not processed. write same
func (self *Mediator) Read() {

TOP:
	index := 0
	try := 0
	self.Mconn.SetReadDeadline(time.Now().Add(time.Minute * 3))

	var data []byte = make([]byte, 2048)

	for index < len(data) {

		n, err := self.Mconn.Read(data[index:])

		if err != nil {
			e, ok := err.(net.Error)
			if !ok || !e.Temporary() || try >= 3 {

				self.Error <- true
			}
			try++
		}
		//收到心跳包
		if data[0] == 'h' && data[1] == 'h' {

			self.Mconn.Write([]byte("hh"))
			goto TOP
		}

		index += n
	}

	self.recv <- data

	goto TOP
}

func (self *Mediator) Write() error {

TOP:
	index := 0
	try := 0
	var data []byte = make([]byte, 2048)

	select {

	case data = <-self.send:

		for index < len(data) {
			n, err := self.Mconn.Write(data[index:])
			if err != nil {

				e, ok := err.(net.Error)

				if !ok || !e.Temporary() || try >= 3 {

					self.Error <- true
				}
				try++
			}
			index += n
		}

	case <-self.writ:
		//Service shutdown Connect
		break

	}

	goto TOP
}

//mediator listen function
func MediatorAccept(conn net.Listener) net.Conn {

	CorU, err := conn.Accept()
	logExit(CorU, err)
	return CorU
}

//exit goroutine
func logExit(conn net.Conn, err error) {

	defer conn.Close()

	if err != nil {
		//forced out of goroutine
		runtime.Goexit()
	}
}

//长连接
func handleConnection(conn net.Conn, timeout int) {

	buffer := make([]byte, 2048)

	for {

		n, err := conn.Read(buffer)

		if err != nil {
			//logErr(conn.RemoteAddr().String(), "Connection Error:", err)
			return
		}

		Data := (buffer[:n])

		msg := make(chan byte, 0)
		postData := make(chan byte, 0)

		//心跳检测
		go HeartBeating(conn, msg, timeout)

		//检测每次Client是否有数据传来
		go GravelChannel(Data, msg)

		//记录日志

	}
}

//根据数据监控，判断是否在设定的时间内发来信息
func HeartBeating(conn net.Conn, msg chan byte, timeout int) {

	select {
	case <-msg:
		//记录日志
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		break
	case <-time.After(time.Second * 5):
		//记录日志
		conn.Close()
	}
}

//数据监控
func GravelChannel(n []byte, msg chan byte) {
	for _, v := range n {
		msg <- v
	}
	close(msg)
}
