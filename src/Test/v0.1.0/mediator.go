//server端，运行在有外网ip的服务器上
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
)

var cport *string = flag.String("cport", "3002", "user访问地址端口")
var mport *string = flag.String("mport", "20012", "与client通讯端口")

//与server相关的conn
type mediator struct {
	conn net.Conn
	er   chan bool
	//未收到心跳包通道
	heart chan bool
	//暂未使用！！！原功能tcp连接已经接通，不在需要心跳包
	disheart bool
	writ     chan bool
	recv     chan []byte
	send     chan []byte
}

//读取mediator过来的数据
func (self *mediator) Read() {
	for {
		//40秒没有数据传输则断开
		self.conn.SetReadDeadline(time.Now().Add(time.Second * 40))
		var recv []byte = make([]byte, 10240)
		n, err := self.conn.Read(recv)

		if err != nil {
			//			if strings.Contains(err.Error(), "timeout") && self.disheart {
			//				fmt.Println("两个tcp已经连接,server不在主动断开")
			//				self.conn.SetReadDeadline(time.Time{})
			//				continue
			//			}
			self.heart <- true
			self.er <- true
			self.writ <- true
			//fmt.Println("长时间未传输信息，或者client已关闭。断开并继续accept新的tcp，", err)
		}
		//收到心跳包hh，原样返回回复
		if recv[0] == 'h' && recv[1] == 'h' {
			self.conn.Write([]byte("hh"))
			continue
		}
		self.recv <- recv[:n]

	}
}

//处理心跳包
//func (self server) cHeart() {
//
//	for {
//		var recv []byte = make([]byte, 2)
//		var chanrecv []byte = make(chan []byte)
//		self.conn.SetReadDeadline(time.Now().Add(time.Second * 30))
//		n, err := self.conn.Read(recv)
//		chanrecv <- recv
//		if err != nil {
//			self.heart <- true
//			fmt.Println("心跳包超时", err)
//			break
//		}
//		if recv[0] == 'h' && recv[1] == 'h' {
//			self.conn.Write([]byte("hh"))
//		}
//
//	}
//}

//把数据发送给mediator
func (self mediator) Write() {

	for {
		var send []byte = make([]byte, 10240)
		select {
		case send = <-self.send:
			self.conn.Write(send)
		case <-self.writ:
			//fmt.Println("写入client进程关闭")
			break

		}
	}

}

//与user相关的conn
type user struct {
	conn net.Conn
	er   chan bool
	writ chan bool
	recv chan []byte
	send chan []byte
}

//读取user过来的数据
func (self user) Read() {
	self.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 800))
	for {
		var recv []byte = make([]byte, 10240)
		n, err := self.conn.Read(recv)
		self.conn.SetReadDeadline(time.Time{})
		if err != nil {

			self.er <- true
			self.writ <- true
			//fmt.Println("读取user失败", err)

			break
		}
		self.recv <- recv[:n]
	}
}

//把数据发送给user
func (self user) Write() {

	for {
		var send []byte = make([]byte, 10240)
		select {
		case send = <-self.send:
			self.conn.Write(send)
		case <-self.writ:
			//fmt.Println("写入user进程关闭")
			break

		}
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	if flag.NFlag() != 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	userPort, _ := strconv.Atoi(*cport)
	mediatorPort, _ := strconv.Atoi(*mport)
	if !(userPort >= 0 && userPort < 65536) {
		fmt.Println("端口设置错误")
		os.Exit(1)
	}
	if !(mediatorPort >= 0 && mediatorPort < 65536) {
		fmt.Println("端口设置错误")
		os.Exit(1)
	}

	//监听端口
	m, err := net.Listen("tcp", ":"+*mport)
	log(err)
	u, err := net.Listen("tcp", ":"+*cport)
	log(err)
	//第一条tcp关闭或者与浏览器建立tcp都要返回重新监听
TOP:
	//监听user链接
	Uconn := make(chan net.Conn)
	go UconnAccept(u, Uconn)
	//一定要先接受Server
	fmt.Println("User Accept Ready..")
	Mconn := MediatorAccept(m)
	fmt.Println("Server Accept Ready..")
	recv := make(chan []byte)
	send := make(chan []byte)
	heart := make(chan bool, 1)
	//1个位置是为了防止两个读取线程一个退出后另一个永远卡住
	er := make(chan bool, 1)
	writ := make(chan bool)
	mediator := &mediator{Mconn, er, heart, false, writ, recv, send}
	go mediator.Read()
	go mediator.Write()

	//这里可能需要处理心跳
	for {
		select {
		case <-mediator.heart:
			goto TOP
		case userconnn := <-Uconn:
			//暂未使用
			mediator.disheart = true
			recv = make(chan []byte)
			send = make(chan []byte)
			//1个位置是为了防止两个读取线程一个退出后另一个永远卡住
			er = make(chan bool, 1)
			writ = make(chan bool)
			user := &user{userconnn, er, writ, recv, send}
			go user.Read()
			go user.Write()
			//当两个socket都创立后进入handle处理
			go handle(mediator, user)
			goto TOP
		}

	}

}

//监听端口函数
func MediatorAccept(con net.Listener) net.Conn {
	CorU, err := con.Accept()
	logExit(err)
	return CorU
}

//在另一个协程中监听端口函数
func UconnAccept(con net.Listener, Uconn chan net.Conn) {
	CorU, err := con.Accept()
	logExit(err)
	Uconn <- CorU
}

//显示错误
func log(err error) {
	if err != nil {
		fmt.Printf("出现错误： %v\n", err)
	}
}

//显示错误并退出
func logExit(err error) {
	if err != nil {
		//fmt.Printf("出现错误，退出线程： %v\n", err)
		runtime.Goexit()
	}
}

//显示错误并关闭链接，退出线程
func logClose(err error, conn net.Conn) {
	if err != nil {
		//fmt.Println("对方已关闭", err)
		runtime.Goexit()
	}
}

//两个socket衔接相关处理
func handle(mediator *mediator, user *user) {
	for {
		var mediatorrecv = make([]byte, 10240)
		var userrecv = make([]byte, 10240)
		select {

		case mediatorrecv = <-mediator.recv:
			user.send <- mediatorrecv
		case userrecv = <-user.recv:
			//fmt.Println("浏览器发来的消息", string(userrecv))
			mediator.send <- userrecv
			//user出现错误，关闭两端socket
		case <-user.er:
			//fmt.Println("user关闭了，关闭server与user")
			mediator.conn.Close()
			user.conn.Close()
			runtime.Goexit()
			//client出现错误，关闭两端socket
		case <-mediator.er:
			//fmt.Println("client关闭了，关闭client与user")
			user.conn.Close()
			mediator.conn.Close()
			runtime.Goexit()
		}
	}
}

// //长连接
// func handleConnection(conn net.Conn, timeout int) {
//
// 	buffer := make([]byte, 2048)
//
// 	for {
//
// 		n, err := conn.Read(buffer)
//
// 		if err != nil {
// 			return
// 		}
//
// 		Data := (buffer[:n])
//
// 		msg := make(chan byte, 0)
// 		postData := make(chan byte, 0)
//
// 		//心跳检测
// 		go HeartBeating(conn, msg, timeout)
//
// 		//检测每次Client是否有数据传来
// 		go GravelChannel(Data, msg)
//
// 		//记录日志
//
// 	}
// }
//
// //根据数据监控，判断是否在设定的时间内发来信息
// func HeartBeating(conn net.Conn, msg chan byte, timeout int) {
//
// 	select {
// 	case <-msg:
// 		//记录日志
// 		conn.SetDeadline(time.Duration(timeout) * time.Second)
// 		break
// 	case <-time.After(time.Second * 5):
// 		//记录日志
// 		conn.Close()
// 	}
// }
//
// //数据监控
// func GravelChannel(n []byte, msg chan byte) {
// 	for _, v := range n {
// 		msg <- n
// 	}
// 	close(msg)
// }
