package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"ziface"
)

/*
链接模块
*/

type Connection struct {
	//当前连接的socket TCP套接字
	Conn *net.TCPConn

	//当前链接的id
	ConnID uint32

	//当前的链接状态
	isClosed bool

	//有router方法了
	// //当前链接所绑定的处理业务方法的API
	// handleAPI ziface.HandleFunc

	//告知当前链接已经退出/停止的 channel (由Reader告知Writer退出)
	ExitChan chan bool

	// //该链接处理的方法Router
	// Router ziface.IRouter
	//V0.6改为Msghandler

	//V0.7新增 无缓冲管道，用于读写Goroutine之间的消息通道
	msgChan chan []byte

	//消息的管理MsgID 和对应的处理业务API关系
	MsgHandler ziface.IMsgHandle
}

func NewConnection(conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		Conn:       conn,
		ConnID:     connID,
		MsgHandler: msgHandler,
		isClosed:   false,
		msgChan:    make(chan []byte),
		ExitChan:   make(chan bool, 1),
	}
	return c
}

func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running]")
	defer fmt.Println("[Reader is exit!],connID =", c.ConnID, "remote addr is", c.RomoteAddr().String()) //写日志
	defer c.Stop()

	for {
		//读取客户端的数据到buf中，最大512字节
		// buf := make([]byte, utils.GlobalObject.MaxPackageSize)
		// _, err := c.Conn.Read(buf)
		// if err != nil {
		// 	fmt.Println("recv buf err", err)
		// 	continue
		// }

		//-----------------方法一------------------
		//创建一个拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg Head 二进制流 8个字节
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConn(), headData); err != nil {
			fmt.Println("read msg head error", err)
			break
		}

		//拆包，得到msgID 和 msgDatalen放在一个message中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
			break
		}

		//根据datalen 再次读取Data,放在msg.Data中
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			if _, err := io.ReadFull(c.GetTCPConn(), data); err != nil {
				fmt.Println("read msg data error", err)
				break
			}
		}

		msg.SetData(data)

		// //----------------方法二-------------------
		// //创建一个拆包解包的对象
		// dp := NewDataPack()

		// //读取客户端的Msg Head 二进制流 8个字节
		// headData := make([]byte, dp.GetHeadLen())
		// if _, err := io.ReadFull(c.GetTCPConn(), headData); err != nil {
		// 	fmt.Println("read msg head error", err)
		// 	break
		// }

		// //拆包，得到msgID 和 msgDatalen放在一个message中
		// msgHead, err := dp.Unpack(headData)
		// if err != nil {
		// 	fmt.Println("unpack error", err)
		// 	break
		// }

		// //根据datalen 再次读取Data,放在msg.Data中
		// msg := msgHead.(*Message) //类型断言，把接口类型转为具体类型
		// if msgHead.GetMsgLen() > 0 {
		// 	msg.Data = make([]byte, msg.GetMsgLen())
		// 	if _, err := io.ReadFull(c.GetTCPConn(), msg.Data); err != nil {
		// 		fmt.Println("read msg data error", err)
		// 		break
		// 	}
		// }

		//得到当前conn数据的request请求数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		//判断是否开启工作池
		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经开启了工作池机制，将消息发送给Worker工作池处理即可
			c.MsgHandler.SendMsgTaskQueue(&req)
		} else {
			//从路由中，找到注册绑定的Conn对应的router调用
			//根据绑定好的MsgID找到对应处理吗api业务 执行
			go c.MsgHandler.DoMsgHandler(&req)
		}

		//从路由中，找到注册绑定的Conn对应的router调用
		//根据绑定好的MsgID 找到对应处理api业务 执行
		//V0.8 迭代掉 go c.MsgHandler.DoMsgHandler(&req)

		// //V0.6迭代掉 执行注册的路由方法
		// go func(request ziface.IRequest) {
		// 	//从路由中，找到注册绑定的Conn对应的router调用
		// 	c.Router.PreHandle(request)
		// 	c.Router.Handle(request)
		// 	c.Router.PostHandle(request)
		// }(&req)

		//在V0.3中改为调用router方法
		// //调用当前链接所绑定的HandleAPI
		// if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
		// 	fmt.Println("ConnID", c.ConnID, "handle is error", err)
		// 	break
		// }
	}
}

/*
写消息的Goroutine,专门发送给客户端消息的模块
*/
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println("[conn Writer exit!]", c.RomoteAddr().String())

	//不断的阻塞等待channel的消息，进行写给客户端
	for {
		select {
		case data := <-c.msgChan:
			//有数据要写回客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error:", err)
				return
			}
		case <-c.ExitChan:
			//代表Reader已经退出，此时Writer也要退出
			return
		}
	}
}

// 启动链接 让当前的链接开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start().. ConnID =", c.ConnID)

	//启动从当前链接的读数据的业务
	go c.StartReader()

	//启动从当前链接写数据的业务
	go c.StartWriter()
}

// 停止链接 结束当前链接的工作
func (c *Connection) Stop() {
	fmt.Println("Conn Stop().. ConnID =", c.ConnID) //打印日志

	//如果当前链接已经关闭
	if c.isClosed == true {
		return
	}

	c.isClosed = true

	//关闭socket链接
	c.Conn.Close()

	//告知Writer关闭
	c.ExitChan <- true

	//回收资源
	close(c.ExitChan)
	close(c.msgChan)

}

// 获取当前链接的绑定socket conn
func (c *Connection) GetTCPConn() *net.TCPConn {
	return c.Conn
}

// 获取当前链接模块的链接id
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端的TCP状态 IP Port
func (c *Connection) RomoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 发送数据，将数据发送给远程的客户端
// 提供一个SendMsg方法 将我们要发送给客户端的数据，先进行封包，再发送
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}

	//将data进行封包 MsgDatalen + Msg + Data
	dp := NewDataPack()

	//创建一个Message

	binarymsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg")
	}

	//V0.7迭代 不通过conn发发回客户端 而是通过startwriter发回客户端
	// //将数据发送给客户端
	// if _, err := c.Conn.Write(binarymsg); err != nil {
	// 	fmt.Println("Write msg id", msgId, "error/:", err)
	// 	return errors.New("conn Write error")
	// }

	//把数据发给管道 管道Writer发给客户端
	c.msgChan <- binarymsg

	return nil
}
