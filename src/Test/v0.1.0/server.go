//client端，运行在家里有网站的电脑中
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var host *string = flag.String("host", "127.0.0.1", "请输入转发服务器ip")
var mport *string = flag.String("mport", "20012", "转发端口")
var sport *string = flag.String("sport", "80", "服务端口")

//与service相关的conn
type services struct {
	conn net.Conn
	er   chan bool
	writ chan bool
	recv chan []byte
	send chan []byte
}

//读取service过来的数据
func (self services) read() {

	for {
		var recv []byte = make([]byte, 10240)
		n, err := self.conn.Read(recv)
		if err != nil {

			self.writ <- true
			self.er <- true
			//fmt.Println("读取browser失败", err)
			break
		}
		self.recv <- recv[:n]

	}
}

//把数据发送给service
func (self services) write() {

	var send []byte = make([]byte, 10240)

	for {
		select {

		case send = <-self.send:
			self.conn.Write(send)
		case <-self.writ:
			goto EXIT
		}
	}

EXIT:
}

//与mediator相关的conn
type mediator struct {
	conn net.Conn
	er   chan bool
	writ chan bool
	recv chan []byte
	send chan []byte
}

//读取mediator过来的数据
func (self *mediator) read() {
	//isheart与timeout共同判断是不是自己设定的SetReadDeadline
	var isheart bool = false
	//20秒发一次心跳包
	self.conn.SetReadDeadline(time.Now().Add(time.Second * 20))
	for {
		var recv []byte = make([]byte, 10240)
		n, err := self.conn.Read(recv)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") && !isheart {
				//fmt.Println("发送心跳包")
				self.conn.Write([]byte("hh"))
				//4秒时间收心跳包
				self.conn.SetReadDeadline(time.Now().Add(time.Second * 4))
				isheart = true
				continue
			}
			//浏览器有可能连接上不发消息就断开，此时就发一个0，为了与服务器一直有一条tcp通路
			self.recv <- []byte("0")
			self.er <- true
			self.writ <- true
			//fmt.Println("没收到心跳包或者server关闭，关闭此条tcp", err)
			break
		}
		//收到心跳包
		if recv[0] == 'h' && recv[1] == 'h' {
			//fmt.Println("收到心跳包")
			self.conn.SetReadDeadline(time.Now().Add(time.Second * 20))
			isheart = false
			continue
		}
		self.recv <- recv[:n]
	}
}

//把数据发送给mediator
func (self mediator) write() {

	for {
		var send []byte = make([]byte, 10240)

		select {
		case send = <-self.send:
			self.conn.Write(send)
		case <-self.writ:
			//fmt.Println("写入server进程关闭")
			break
		}

	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if flag.NFlag() != 3 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	servicePort, _ := strconv.Atoi(*sport)
	mediatorPort, _ := strconv.Atoi(*mport)
	if !(servicePort >= 0 && servicePort < 65536) {
		fmt.Println("端口设置错误")
		os.Exit(1)
	}
	if !(mediatorPort >= 0 && mediatorPort < 65536) {
		fmt.Println("端口设置错误")
		os.Exit(1)
	}
	targetHost := net.JoinHostPort(*host, *mport)
	for {
		//链接端口
		Mconn := dail(targetHost)
		recv := make(chan []byte)
		send := make(chan []byte)
		//1个位置是为了防止两个读取线程一个退出后另一个永远卡住
		er := make(chan bool, 1)
		writ := make(chan bool)
		next := make(chan bool)
		mediator := &mediator{Mconn, er, writ, recv, send}
		go mediator.read()
		go mediator.write()
		go handle(mediator, next)
		<-next
	}

}

//显示错误并退出
func logExit(err error) {
	if err != nil {
		fmt.Printf("出现错误，退出线程： %v\n", err)
		runtime.Goexit()
	}
}

//链接端口
func dail(hostport string) net.Conn {
	conn, err := net.Dial("tcp", hostport)
	logExit(err)
	return conn
}

//两个socket衔接相关处理
func handle(mediator *mediator, next chan bool) {
	var mediatorrecv = make([]byte, 10240)
	//阻塞这里等待server传来数据再链接browser
	fmt.Println("等待mediator发来消息")
	mediatorrecv = <-mediator.recv
	//连接上，下一个tcp连上服务器
	next <- true
	//fmt.Println("开始新的tcp链接，发来的消息是：", string(serverrecv))
	var service *services
	//server发来数据，链接本地80端口
	serviceconn := dail("127.0.0.1:" + *sport)
	recv := make(chan []byte)
	send := make(chan []byte)
	er := make(chan bool, 1)
	writ := make(chan bool)
	service = &services{serviceconn, er, writ, recv, send}
	go service.read()
	go service.write()
	service.send <- mediatorrecv

	for {
		var mediatorrecv = make([]byte, 10240)
		var servicerecv = make([]byte, 10240)
		select {
		case mediatorrecv = <-mediator.recv:
			if mediatorrecv[0] != '0' {

				service.send <- mediatorrecv
			}

		case servicerecv = <-service.recv:
			mediator.send <- servicerecv
		case <-mediator.er:
			//fmt.Println("mediator关闭了，关闭service与mediator")
			mediator.conn.Close()
			service.conn.Close()
			runtime.Goexit()
		case <-service.er:
			//fmt.Println("browse关闭了，关闭server与browse")
			mediator.conn.Close()
			service.conn.Close()
			runtime.Goexit()
		}
	}
}
