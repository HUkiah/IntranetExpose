package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

//与server相关的conn
type Server struct {
	Sconn  net.Conn
	IsDead chan bool
	Error  chan bool
	Recv   chan []byte
	Send   chan []byte
}

var sport *string = flag.String("sport", "20012", "与client通讯端口")

func Log(v ...interface{}) {
	log.Println(v...)
}

func (self *Server) Read() {

	data := make([]byte, 10240)

TOP:
	try := 0

	self.Sconn.SetReadDeadline(time.Now().Add(time.Second * 420))

	for {

		n, err := self.Sconn.Read(data)

		if err != nil {
			Log(self.Sconn.RemoteAddr().String(), " connection error: ", err)
			// if strings.Contains(err.Error(), "EOF") {
			// 	self.IsDead <- true
			//   return
			// }

			//不是一个网络错误，临时错误，已经试了3次
			if strings.Contains(err.Error(), "timeout") && try <= 5 {
				//取消超时
				self.Sconn.SetReadDeadline(time.Time{})
				//如果因为网络问题，延长等待30s。没有导致没有收到，则发送3次心跳
				if try >= 3 {
					//try 3 次后，发送3心跳包
					try++
					self.Send <- []byte("hh")
					self.Sconn.SetReadDeadline(time.Now().Add(20 * time.Second))
					continue
				}
				try++
				self.Sconn.SetReadDeadline(time.Now().Add(20 * time.Second))
				continue
			}

			self.IsDead <- true
			return
		}

		Log(self.Sconn.RemoteAddr().String(), "Receive :\n", string(data[:n]))

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

func (self *Server) Write() {

	data := make([]byte, 10240)

	for {

		select {
		case data = <-self.Send:
			Log(self.Sconn.RemoteAddr().String(), "Send: \n", string(data))
			self.Sconn.Write(data)
		case <-self.IsDead:
			//Do Something
			self.Error <- true

		}

	}

}

func Accpet(con net.Listener, Sconn chan net.Conn) {
	Con, err := con.Accept()
	Log(err)
	Sconn <- Con
}

func main() {

	flag.Parse()
	if flag.NFlag() != 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	//sPort, _ := strconv.Atoi(*sport)

	s, err := net.Listen("tcp", ":"+*sport)
	Log(err)
TOP:
	Sconn, err := s.Accept()
	Log(err)
	//conn := make(chan net.Conn, 0)

	// go Accpet(s, conn)
	//
	// Sconn := <-conn

	server := &Server{Sconn, make(chan bool), make(chan bool), make(chan []byte), make(chan []byte)}

	go server.Read()
	go server.Write()

	<-server.Error
	server.Sconn.Close()
	Log("Sconn close")
	goto TOP
}
