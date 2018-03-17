package ServiceProvider

import (

  "net"

)

type Service struct {

  Sconn net.Conn
  Error chan bool
  writ  chan bool
  recv  chan []byte
  send  chan []byte
}

//
func (self *Service) Read() {

  for {
    data := make([]byte, 2048)

    n,err := self.Sconn.Read(data)

    if err!=nil {
      self.Error <- true

    }

    self.recv <- data[:n]

  }

}

//
func (self *Service) Write() {

  for {

    data := make([]byte, 2048)

    select {

      case data = <-self.send:
        self.Sconn.Write(data)

    }
  }
}
