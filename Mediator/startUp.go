package Mediator

import (

  "runtime"
  "flag"
  "fmt"
  "os"
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


func Atoi(str string) (result int, errorMsg string) {

   if port, _ := strconv.Atoi(str); !(port >= 1000 && port < 65536) {

     dData := Port{
       port:port,
     }

     errorMsg = dData.Error()
     return
   }else{
     return port , ""
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
  if ServicePort, errorMsg := Atoi(*sport); errorMsg != "" {
      fmt.Println("errorMsg is: ", errorMsg)
  }
  if MediatorPort, errorMsg := Atoi(*mport); errorMsg != "" {
      fmt.Println("errorMsg is: ", errorMsg)
  }
  if match,_ := regexp.MatchString(`^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`, *host);match{
    fmt.Println("host IP right")
  }else{
    fmt.Println("IP address Format Error")
  }




}
