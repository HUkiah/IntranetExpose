package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type Client struct {
	Cconn  net.Conn
	IsDead chan bool
	Error  chan bool
	Recv   chan []byte
	Send   chan []byte
}

var host *string = flag.String("host", "127.0.0.1", "请输入服务器ip")
var sport *string = flag.String("sport", "20012", "服务端口")

func Log(v ...interface{}) {
	log.Println(v...)
}

func (self *Client) Read() {

	data := make([]byte, 10240)

TOP:
	try := 0

	self.Cconn.SetReadDeadline(time.Now().Add(time.Second * 6))

	for {
		n, err := self.Cconn.Read(data)

		//记录是否断开
		if err != nil {
			Log(self.Cconn.RemoteAddr().String(), " connection error: ", err)

			//server 关闭
			// if strings.Contains(err.Error(), "EOF") && try >= 5 {
			//   self.IsDead <- true
			// 	return
			// }

			//不是一个网络错误，临时错误，已经试了3次
			if strings.Contains(err.Error(), "timeout") && try <= 5 {
				//取消超时
				self.Cconn.SetReadDeadline(time.Time{})
				//如果因为网络问题，延长等待30s。没有导致没有收到，则发送3次心跳
				if try >= 3 {
					//try 3 次后，发送3心跳包
					try++
					self.Send <- []byte("xx")
					self.Cconn.SetReadDeadline(time.Now().Add(2 * time.Second))
					continue
				}
				try++
				self.Cconn.SetReadDeadline(time.Now().Add(1 * time.Second))
				continue
			}

			self.IsDead <- true
			return
		}

		Log(self.Cconn.RemoteAddr().String(), "Receive:\n", string(data[:n]))

		// if data[0] == 'x' && data[1] == 'x' {
		// 	self.IsDead <- true
		// 	runtime.Goexit()
		// } else {
		//
		// 	try++
		// 	self.Send <- []byte(strconv.Itoa(try))
		// }

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

func (self *Client) Write() {
	data := make([]byte, 10240)

	for {

		select {
		case data = <-self.Send:
			Log(self.Cconn.RemoteAddr().String(), "Send:\n", string(data))
			self.Cconn.Write(data)
		case <-self.IsDead:
			//Do Something
			self.Error <- true
			//return
		}

	}
}

func main() {

	flag.Parse()
	if flag.NFlag() != 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	Cconn, err := net.Dial("tcp", net.JoinHostPort(*host, *sport))
	Log(err)

	client := &Client{Cconn, make(chan bool, 1), make(chan bool), make(chan []byte, 0), make(chan []byte, 0)}

	go client.Read()
	go client.Write()

	// client.Send <- []byte("0")
	//
	// go func() {
	// 	time.Sleep(time.Microsecond * 50)
	//
	// 	client.Send <- []byte("xx")
	// }()

	select {

	case <-client.Error:
		client.Cconn.Close()
		Log("Server EOF")

	}

}
