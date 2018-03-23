package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

//定义客户端结构体
type Client struct {
	C chan string
	Username string
	CliAddr string
}

//在线用户map
var onlineMap map[string]Client

//通信管道
var message = make(chan string)

//处理用户连接
func HandConn(conn net.Conn){
	defer conn.Close()

	//获取客户端地址
	CliAddr := conn.RemoteAddr().String()
	//创建结构体
	cli := Client{make(chan string), CliAddr, CliAddr}
	//加入到map
	onlineMap[CliAddr] = cli

	//新开一个协程 专门给当前客户端发信息
	go WrteMsgTOClient(cli, conn)

	//广播某个人在线
	message <- MakeMsg(cli, "login")

	//提示 我是
	cli.C <- MakeMsg(cli, "我是：" + CliAddr)

	isQuit := make(chan bool) //对方是否主动退出
	hasData := make(chan bool) //有数据

	//新建一个协程，接收用户发过来的数据
	go func() {
		buf := make([]byte, 2048)

		for{
			n, err := conn.Read(buf)
			if n == 0 {
				//对方断开或者出问题
				isQuit <- true
				fmt.Println("conn.Read err = ", err)
				return
			}
			//转发此内容
			msg := string(buf[:n-1]) // nc测试会多一个换行

			if len(msg) == 3 && msg == "who"{
				//遍历map，给当前用户发送成员
				conn.Write([]byte("user list:\n"))
				for _,tmp := range onlineMap {
					msg = tmp.CliAddr + ":" + tmp.Username + "\n"
					conn.Write([]byte(msg))
				}
			}else if  len(msg) >= 8 && msg[:6] == "rename"{
				//rename|mike
				name := strings.Split(msg, "|")[1]
				cli.Username = name
				onlineMap[CliAddr] = cli
				conn.Write([]byte("改名成功\n"))
			}else if len(msg) == 4 && msg == "exit" {
				isQuit <- true
			}else{
				message <- MakeMsg(cli, msg)
			}

			hasData <- true
		}
	}()

	for{
		select {
			//检查
		case <- isQuit:
			delete(onlineMap, CliAddr) // 用户移除
			message <- MakeMsg(cli, "login out") //广播下线
			return //断掉当前链接
		case <- hasData:
		case <-time.After(10 * time.Second): // 60s后
			//超时处理
			delete(onlineMap, CliAddr) //从当前下线
			message <- MakeMsg(cli, "time out") //广播下线
			return //断掉当前链接
		}
	}
}

func MakeMsg(cli Client, msg string) string{
	return "[" + cli.CliAddr + "] " + cli.Username + ": " + msg
}

func WrteMsgTOClient(cli Client, conn net.Conn){
	for{
		for msg := range cli.C{
			conn.Write([]byte(msg + "\n"))
		}

	}
}

func Messager(){
	//给map分配空间
	onlineMap = make(map[string]Client)
	for{
		msg := <- message //没有消息前 是阻塞的
		//遍历发消息
		for _, cli := range onlineMap {
			cli.C <- msg
		}
	}
}

func main() {
	//监听
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("net.Listen err = ", err)
		return
	}
	defer listener.Close()

	//新开协程 转发消息 只要有消息， 给map发消息
	go Messager()

	//循环接收用户请求
	for{
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err = ", err)
			continue // 可能下一个用户是链接的 所以不能用return
		}

		//处理用户连接
		go HandConn(conn)
	}
}
