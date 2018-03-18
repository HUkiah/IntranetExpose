package main

import (
	"Mediator/MediatorService"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
)

var mport *string = flag.String("mport", "20012", "代理端口")
var cport *string = flag.String("cport", "3000", "服务端口")

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	if flag.NFlag() != 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Port Overflow verification
	if errorMsg := MediatorService.Atoi(*cport); errorMsg != "" {
		fmt.Println("errorMsg is: ", errorMsg)
	}
	if errorMsg := MediatorService.Atoi(*mport); errorMsg != "" {
		fmt.Println("errorMsg is: ", errorMsg)
	}

	//监听端口
	m, err := net.Listen("tcp", ":"+*mport)
	if err != nil {
		fmt.Printf("Listen Mediator Port Error： %v\n", err)
	}
	c, err := net.Listen("tcp", ":"+*cport)
	if err != nil {
		fmt.Printf("Listen Client Port Error: %v\n", err)
	}

INIT:
	//开启协程监听Client链接
	Cconn := make(chan net.Conn)
	go MediatorService.CconnAccept(c, Cconn)

	//等待初始化ServiceProvider
	Mconn := MediatorService.MediatorAccept(m)

	mediator := &MediatorService.Mediator{Mconn, make(chan bool, 0), make(chan bool, 1), make(chan bool), make(chan []byte, 0), make(chan []byte, 0)}

	go mediator.Read()
	go mediator.Write()

	for {
		select {

		case <-mediator.Heart:
			goto INIT
			//等待client connect
		case Clientconn := <-Cconn:
			client := &MediatorService.Client{Clientconn, make(chan bool, 1), make(chan bool), make(chan []byte, 0), make(chan []byte, 0)}
			go client.Read()
			go client.Write()
			go handler(mediator, client)
		}
	}

}

func handler(mediator *MediatorService.Mediator, client *MediatorService.Client) {

	fmt.Println("Handler...")
	clientrecv := make([]byte, 2048)
	mediatorrecv := make([]byte, 2048)

TOP:

	select {

	case mediatorrecv = <-mediator.Recv:
		client.Send <- mediatorrecv

	case clientrecv = <-client.Recv:
		mediator.Send <- clientrecv

	case <-mediator.Error:
		mediator.Mconn.Close()
		client.Cconn.Close()
		runtime.Goexit()
	case <-client.Error:
		mediator.Mconn.Close()
		client.Cconn.Close()
		runtime.Goexit()

	}

	goto TOP

}
