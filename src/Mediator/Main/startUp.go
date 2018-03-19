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

MINIT:
	//等待初始化ServiceProvider
	Mconn := MediatorService.MediatorAccept(m)

	mediator := &MediatorService.Mediator{Mconn, make(chan bool), make(chan bool), make(chan []byte), make(chan []byte)}

	go mediator.Read()
	go mediator.Write()

CINIT:
	//开启协程监听Client链接
	Cconn := make(chan net.Conn)
	go MediatorService.CconnAccept(c, Cconn)

	var client *MediatorService.Client
	var clientStatus = make(chan bool)

	for {
		select {
		case <-mediator.Error:
			mediator.Mconn.Close()
			MediatorService.Log("Mconn close")
			goto MINIT
			//等待client connect
		case Clientconn := <-Cconn:
			client = &MediatorService.Client{Clientconn, make(chan bool), make(chan bool), make(chan []byte), make(chan []byte)}
			go client.Read()
			go client.Write()
			go handler(mediator, client, clientStatus)
		case <-clientStatus:
			goto CINIT
		}
	}

}

func handler(mediator *MediatorService.Mediator, client *MediatorService.Client, clientStatus chan bool) {

	MediatorService.Log("Handler...")
	clientrecv := make([]byte, 10240)
	mediatorrecv := make([]byte, 10240)

	for {
		select {

		case mediatorrecv = <-mediator.Recv:
			client.Send <- mediatorrecv

		case clientrecv = <-client.Recv:
			mediator.Send <- clientrecv

		case <-client.Error:
			client.Cconn.Close()
			MediatorService.Log("Cconn close")
			clientStatus <- true
		}
	}

}
