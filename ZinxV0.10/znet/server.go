package znet

import (
	"fmt"
	"net"
	"utils"
	"ziface"
)

// IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//服务器名称
	Name string
	//服务器绑定的IP版本
	IPVersion string
	//服务器监听的ip
	IP string
	//服务器监听的端口
	Port int

	//给当前的Server添加一个router，server注册的链接对应的处理业务
	//以后一个服务器可以绑定多个router
	//Router ziface.IRouter

	//V0.6 当前server的消息管理模块，用来绑定MsgID和对应的处理业务API关系
	MsgHandler ziface.IMsgHandle
	//该server的连接管理器
	ConnMgr ziface.IConnManager

	//该Server创建连接之后自动调用Hook函数——OnConnStart 函数指针
	OnConnStart func(conn ziface.IConnect)
	//该Server销毁连接之前自动调用的Hook函数——OnConnStop 函数指针
	OnConnStop func(conn ziface.IConnect)
}

// // 定义当前客户端链接的所绑定的handle api，（目前这个handle是写死的，以后优化应该由用户自定义handle方法）
// func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
// 	//回显的业务

// 	fmt.Println("[Conn Handle] CallBackToClient...")
// 	if _, err := conn.Write(data[:cnt]); err != nil {
// 		fmt.Println("write back buf err", err)
// 		return errors.New("CallBackToClient error")
// 	}

// 	return nil
// }

// 启动服务器
func (s *Server) Start() {
	fmt.Printf("[Zinx]Server Name : %s,listenner at IP : %s,Port:%d is strarting\n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[Zinx] Version %s,MaxConn:%d,MaxPackeetSize:%d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPackageSize)
	fmt.Printf("[Start]Server Listener at IP:%s,Port %d, is starting\n", s.IP, s.Port)

	//防止阻塞 异步
	go func() {
		//V0.8 0
		s.MsgHandler.StartWorkPool()

		//1、获取一个TCP的addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addt err:", err)
			return
		}
		//2、监听这个服务器的地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		fmt.Println("start Zinx server succ", s.Name, "succ,Listening...")
		var cid uint32
		cid = 0

		//3、阻塞的等待客户端连接，处理客户端连接业务(读写)
		for {
			//如果有客户端连接过来，阻塞会返回
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			//设置最大连接个数的判断，如果超过最大连接，那么则关闭此新的连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				//TODO 给客户端相应一个超出最大连接的错误包
				fmt.Println("======>Too Many Connection MaxConn = ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}

			//将处理新链接的业务方法 和 conn 进行绑定，得到我们的链接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandler)
			cid++

			//启动当前的新链接处理业务
			go dealConn.Start()

			//用connection模块封装v0.2
			// //已经与客户端建立链接，做一些业务，做一个最基本的最大512字节长度的回显业务
			// go func() {
			// 	for {
			// 		buf := make([]byte, 512)
			// 		cnt, err := conn.Read(buf)
			// 		if err != nil {
			// 			fmt.Println("recv bff err", err)
			// 			continue
			// 		}

			// 		fmt.Printf("recv client buf %s,cnt %d\n", buf, cnt)

			// 		//回显功能
			// 		if _, err := conn.Write(buf[:cnt]); err != nil {
			// 			fmt.Println("write back buf err", err)
			// 			continue
			// 		}
			// 	}
			// }()
		}
	}()

}

// 停止服务器
func (s *Server) Stop() {
	//将一些服务器的资源、状态或者一些已经开辟的连接信息 进行停止或者回收
	fmt.Println("[STOP] Zinx server name", s.Name)
	s.ConnMgr.ClearConn()
}

// 运行服务器
func (s *Server) Serve() {
	//启动server的服务功能
	s.Start() //异步

	//TODO 做一些启动服务器之后的额外业务

	//阻塞状态
	select {}
}

// 路由功能，给当前的服务注册一个路由方法，供客户端的链接处理使用
func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	//s.Router = router
	//V0.6加入消息管理模块
	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("Add Router Succ!!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

/*
初始化Server模块的方法
*/
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		//Router:    nil,
		MsgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}

	return s
}

// 注册OnConnStart 钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnect)) {
	s.OnConnStart = hookFunc
}

// 注册OnConnStop 钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnect)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStart 钩子函数的方法
func (s *Server) CallOnConnStart(conn ziface.IConnect) {
	if s.OnConnStart != nil {
		fmt.Println("————> Call OnConnStart()...")
		s.OnConnStart(conn)
	}
}

// 调用OnConnStop 钩子函数的方法
func (s *Server) CallOnConnStop(conn ziface.IConnect) {
	if s.OnConnStop != nil {
		fmt.Println("————> Call OnConnStop()...")
		s.OnConnStop(conn)
	}
}
