package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"znet"
)

//   模拟客户端

func main() {
	fmt.Println("client start...")

	time.Sleep(1 * time.Second)

	//1、直接连接远程服务器，得到一个conn连接
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client0 start err,exit!", err)
		return
	}

	for {
		//----------为解决粘包，改写发送消息业务-----------
		// //2、链接调用Write 写数据
		// _, err := conn.Write([]byte("Hello Zinx V0.4.."))
		// if err != nil {
		// 	fmt.Println("write conn err", err)
		// 	return
		// }

		// buf := make([]byte, 512)
		// cnt, err := conn.Read(buf)
		// if err != nil {
		// 	fmt.Printf("read buf error")
		// 	return
		// }

		// fmt.Printf("server call back :%s,cnt=%d\n", buf, cnt)

		//--------发送封包的message消息 MsgID=0
		dp := znet.NewDataPack()
		binaryMsg, err := dp.Pack(znet.NewMsgPackage(0, []byte("Zinx client0 Test Message")))
		if err != nil {
			fmt.Println("Pack error:", err)
			return
		}
		if _, err := conn.Write(binaryMsg); err != nil {
			fmt.Println("write error")
			return
		}

		//----------读取服务器发来的数据
		//服务器就应该给我们回复一个message，MsgID=1 ping...ping...ping
		//拆包
		//1、先读取流中的head部分，得到id和datalen

		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head error", err)
			break
		}
		//将二进制的head拆包到msg结构体中
		//方法一
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client unpack msghead", err)
			break
		}
		if msgHead.GetMsgLen() > 0 {
			//2、再根据Datalen进行第二次读取，将data读出来
			msg := msgHead.(*znet.Message) //类型断言
			msg.Data = make([]byte, msg.GetMsgLen())

			if _, err := io.ReadFull(conn, msg.Data); err != nil {
				fmt.Println("read msg data error", err)
				return
			}
			fmt.Println("————>Recv Server Msg: ID:", msg.Id, "len:", msg.DataLen, "data=", string(msg.Data))
		}

		//cpu阻塞 防止cpu不断判断 跑满了
		time.Sleep(1 * time.Second)
	}

}
