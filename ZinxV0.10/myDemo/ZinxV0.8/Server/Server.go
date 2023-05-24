package main

import (
	"fmt"
	"ziface"
	"znet"
)

/*
   基于Zinx框架来开发的 服务器端应用程序
*/

// ping test自定义路由
type PingRouter struct {
	znet.BaseRouter
}

// // test PreHandle
// func (this *PingRouter) PreHandle(request ziface.IRequest) {
// 	fmt.Println("Call Router PreHandle..")
// 	_, err := request.GetConnection().GetTCPConn().Write([]byte("before ping...\n"))
// 	if err != nil {
// 		fmt.Println("call back bafore ping error")
// 	}
// }

// teat Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call Router Handle..")
	//request中加上message后重写router
	// _, err := request.GetConnection().GetTCPConn().Write([]byte("ping...ping...ping...\n"))
	// if err != nil {
	// 	fmt.Println("call back ping...ping...ping... error")
	// }
	//先读取客户端的数据，再回写ping...ping...ping...业务
	fmt.Println("recv from client:msgID =", request.GetMsgId(),
		",data = ", string(request.GetData()))
	//回写
	err := request.GetConnection().SendMsg(200, []byte("ping...ping...ping..."))
	if err != nil {
		fmt.Println(err)
	}
}

// // test PostHandle
// func (this *PingRouter) PostHandle(request ziface.IRequest) {
// 	fmt.Println("Call Router PostHandle..")
// 	_, err := request.GetConnection().GetTCPConn().Write([]byte("after ping...\n"))
// 	if err != nil {
// 		fmt.Println("call back after ping... error")
// 	}
// }

// hello Zinx test自定义路由
type HelloZinxRouter struct {
	znet.BaseRouter
}

// teat Handle
func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle..")
	//先读取客户端的数据，再回写ping...ping...ping...业务
	fmt.Println("recv from client:msgID =", request.GetMsgId(),
		",data = ", string(request.GetData()))
	//回写
	err := request.GetConnection().SendMsg(201, []byte("Hello! Welcome to Zinx!"))
	if err != nil {
		fmt.Println(err)
	}
}

// 创建连接之后执行的钩子函数
func DoConnectionBegin(conn ziface.IConnect) {
	fmt.Println("==> DoConnectionBegin is Called ...")
	if err := conn.SendMsg(202, []byte("DoConnection BEGIN")); err != nil {
		fmt.Println(err)
	}

	//给当前连接设置一些属性
	fmt.Println("Set conn property...")
	conn.SetProperty("Name", "PurpleGinsburg")
	conn.SetProperty("Github", "https://github.com/PurpleGinsburg")
}

// 连接断开之前的需要执行的函数
func DoConnectionLost(conn ziface.IConnect) {
	fmt.Println("==> DoConnectionLost is Called ...")
	fmt.Println("conn ID = ", conn.GetConnID(), "is Lost ...")

	//获取连接属性
	if name, err := conn.GetProperty("Name"); err == nil {
		fmt.Println("Name = ", name)
	}

	if github, err := conn.GetProperty("Github"); err == nil {
		fmt.Println("github = ", github)
	}
}

func main() {
	//1.创建一个Server句柄
	s := znet.NewServer("[zinx V0.10]")

	//2.注册连接的Hook钩子函数
	s.SetOnConnStart(DoConnectionBegin)
	s.SetOnConnStop(DoConnectionLost)

	//3.给当前Zinx框架添加自定义的router
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})

	//4.启动server
	s.Serve()
}
