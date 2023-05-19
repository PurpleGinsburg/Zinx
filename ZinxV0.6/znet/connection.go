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

	//告知当前链接已经退出/停止的 channel
	ExitChan chan bool

	// //该链接处理的方法Router
	// Router ziface.IRouter
	//V0.6改为Msghandler

	//消息的管理MsgID 和对应的处理业务API关系
	MsgHandler ziface.IMsgHandle
}

func NewConnection(conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		Conn:       conn,
		ConnID:     connID,
		MsgHandler: msgHandler,
		isClosed:   false,
		ExitChan:   make(chan bool, 1),
	}
	return c
}

func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running...")
	defer fmt.Println("connID =", c.ConnID, "Reader is exit,remote addr is", c.RomoteAddr().String()) //写日志
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

		//从路由中，找到注册绑定的Conn对应的router调用
		//根据绑定好的MsgID 找到对应处理api业务 执行
		go c.MsgHandler.DoMsgHandler(&req)

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

// 启动链接 让当前的链接开始工作
func (c *Connection) Start() {
	fmt.Println("Conn Start().. ConnID =", c.ConnID)

	//启动从当前链接的读数据的业务
	go c.StartReader()

	//TODO 启动从当前链接写数据的业务
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

	//回收资源
	close(c.ExitChan)

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

	//将数据发送给客户端
	if _, err := c.Conn.Write(binarymsg); err != nil {
		fmt.Println("Write msg id", msgId, "error/:", err)
		return errors.New("conn Write error")
	}

	return nil
}
