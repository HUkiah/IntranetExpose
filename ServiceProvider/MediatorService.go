package ServiceProvider

import (

  "net"

)

type Mediator struct {

  Mconn net.Conn
  Error chan bool
  writ  chan bool
  recv  chan []byte
  send  chan []byte
}


//
func (self *Mediator) Read() {


}

//
func (self *Mediator) Write() {


}
