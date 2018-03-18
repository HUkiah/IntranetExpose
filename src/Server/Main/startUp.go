package main

import (
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

TOP:
	//Init Mediator Service
	Mconn, err := net.Dial("tcp", net.JoinHostPort(*host, *mport))
	if err != nil {
		fmt.Sprintf("net.Dial %s Error!", *host+*mport)
	}

	mediator := &ServiceProvider.Mediator{Mconn, make(chan bool, 1), make(chan bool), make(chan []byte, 0), make(chan []byte, 0)}

	go mediator.Read()
	go mediator.Write()

	next := make(chan bool)
	go handler(mediator, next)
	<-next
	goto TOP

}

func handler(mediator *ServiceProvider.Mediator, next chan bool) {

	servicerecv := make([]byte, 10240)
	mediatorrecv := make([]byte, 10240)

	//Init Local services
	Sconn, error := net.Dial("tcp", net.JoinHostPort("localhost", *sport))
	if error != nil {
		fmt.Sprintf("net.Dial %s Error!", *host+*mport)
	}

	service := &ServiceProvider.Service{Sconn, make(chan bool, 1), make(chan bool), make(chan []byte, 0), make(chan []byte, 0)}

	go service.Read()
	go service.Write()

	//mediator msg
	mediatorrecv = <-mediator.Recv
	service.Send <- mediatorrecv
	next <- true

TOP:

	select {

	case mediatorrecv = <-mediator.Recv:
		service.Send <- mediatorrecv
		fmt.Println("Remote Mediator->Server")
	case servicerecv = <-service.Recv:
		mediator.Send <- servicerecv
		fmt.Println("Server->Mediator")

	case <-mediator.Error:
		mediator.Mconn.Close()
		service.Sconn.Close()
		runtime.Goexit()
	case <-service.Error:
		mediator.Mconn.Close()
		service.Sconn.Close()
		runtime.Goexit()

	}

	goto TOP

}
