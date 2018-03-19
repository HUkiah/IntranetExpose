package main

import (
	"Mediator/MediatorService"
	"Server/ServiceProvider"
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
	"runtime"
)

var host *string = flag.String("host", "127.0.0.1", "请输入代理服务器ip")
var mport *string = flag.String("mport", "20012", "代理端口")
var sport *string = flag.String("sport", "80", "服务端口")

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	if flag.NFlag() != 3 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Port Overflow
	if errorMsg := ServiceProvider.Atoi(*sport); errorMsg != "" {
		fmt.Println("errorMsg is: ", errorMsg)
	}
	if errorMsg := ServiceProvider.Atoi(*mport); errorMsg != "" {
		fmt.Println("errorMsg is: ", errorMsg)
	}

	if match, _ := regexp.MatchString(`^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`, *host); !match {
		fmt.Println("IP address Format Error")
	}

MINIT:
	//Init Mediator Service
	Mconn, err := net.Dial("tcp", net.JoinHostPort(*host, *mport))
	if err != nil {
		fmt.Sprintf("net.Dial %s Error!", *host+*mport)
	}

	mediator := &ServiceProvider.Mediator{Mconn, make(chan bool), make(chan bool), make(chan []byte), make(chan []byte)}

	go mediator.Read()
	go mediator.Write()

	//	goto TOP

SINIT:
	//Init Local services
	Sconn, error := net.Dial("tcp", net.JoinHostPort("localhost", *sport))
	if error != nil {
		fmt.Sprintf("net.Dial %s Error!", *host+*mport)
	}

	service := &ServiceProvider.Service{Sconn, make(chan bool), make(chan bool), make(chan []byte), make(chan []byte)}

	go service.Read()
	go service.Write()

	go handler(mediator, service)

	//TOP:
	for {
		select {
		case <-mediator.Error:
			mediator.Mconn.Close()
			MediatorService.Log("Mconn close")
			goto MINIT
		case <-service.Error:
			service.Sconn.Close()
			MediatorService.Log("Sconn close")
			goto SINIT

		}
	}

}

func handler(mediator *ServiceProvider.Mediator, service *ServiceProvider.Service) {

	servicerecv := make([]byte, 10240)
	mediatorrecv := make([]byte, 10240)

	for {
		select {

		case mediatorrecv = <-mediator.Recv:
			service.Send <- mediatorrecv
		case servicerecv = <-service.Recv:
			mediator.Send <- servicerecv

		}
	}

}
