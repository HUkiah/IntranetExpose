package ServiceProvider

import (

  "runtime"
  "flag"
  "fmt"
  "os"
  "net"
  "strconv"
  "regexp"
)

var host *string = flag.String("host", "127.0.0.1", "请输入代理服务器ip")
var mport *string = flag.String("mport", "20012", "代理端口")
var sport *string = flag.String("sport", "80", "服务端口")

type Port struct {

  port int

}

func (pt *Port) Error() string {

  strFormat := `you port is %d , Notice that the range is between 1000-65536. `
  return fmt.Sprintf(strFormat,pt.port)
}


func Atoi(str string) (errorMsg string) {

   if port, _ := strconv.Atoi(str); !(port >= 1000 && port < 65536) {

     dData := Port{
       port:port,
     }

     errorMsg = dData.Error()
     return
   }else{
     return ""
   }
}




func main() {

  runtime.GOMAXPROCS(runtime.NumCPU())

  flag.Parse()
  if flag.NFlag() != 3 {
    flag.PrintDefaults()
    os.Exit(1)
  }

  // Port Overflow
  if errorMsg := Atoi(*sport); errorMsg != "" {
      fmt.Println("errorMsg is: ", errorMsg)
  }
  if errorMsg := Atoi(*mport); errorMsg != "" {
      fmt.Println("errorMsg is: ", errorMsg)
  }
  if match,_ := regexp.MatchString(`^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`, *host); !match{
    fmt.Println("IP address Format Error")
  }

  Mconn, err := net.Dial("tcp",net.JoinHostPort(*host, *mport));if err!=nil {
    fmt.Sprintf("net.Dial %s Error!",*host+*mport)
  }

  mediator := &Mediator{Mconn,make(chan bool,1),make(chan bool),make(chan []byte, 0),make(chan []byte, 0)}

  go mediator.Read()
  go mediator.Write()

  mediatorrecv := make([]byte, 2048)

  //wait mediator msg
  mediatorrecv = <-mediator.recv

  //recv mediator msg
  Sconn,error := net.Dial("tcp", net.JoinHostPort("localhost",*sport))
  if error != nil {
    fmt.Sprintf("net.Dial %s Error!",*host+*mport)
  }

  service := &Service{Sconn,make(chan bool,1),make(chan bool),make(chan []byte, 0),make(chan []byte, 0)}

  go service.Read()
  go service.Write()

  service.send <- mediatorrecv

  defer func(){

      mediator.Mconn.Close()
      service.Sconn.Close()
      runtime.Goexit()
    }()

TOP:
  servicerecv := make([]byte, 2048)

  select {

  case mediatorrecv = <- mediator.recv:
      if mediatorrecv[0] != '0' {
        service.send <- mediatorrecv
      }

  case servicerecv = <- service.recv:
        mediator.send <- servicerecv

  case <-mediator.Error:
  case <-service.Error:
        break
  }

goto TOP
}
