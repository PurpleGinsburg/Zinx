package znet

import (
	"errors"
	"fmt"
	"sync"
	"ziface"
)

/*
   连接管理模块
*/

type ConnManger struct {
	connections map[uint32]ziface.IConnect //管理的连接集合
	connLock    sync.RWMutex               //保护连接集合的互斥锁
}

// 创建当前连接的方法
func NewConnManager() *ConnManger {
	return &ConnManger{
		connections: make(map[uint32]ziface.IConnect),
		//锁不用初始化 开箱即用
	}
}

// 添加连接
func (connMgr *ConnManger) Add(conn ziface.IConnect) {
	//保护共享资源map，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//将conn加入到ConnManger中
	connMgr.connections[conn.GetConnID()] = conn

	fmt.Println("connID =", conn.GetConnID(), "connection add to ConnManager successfully : conn num = ", connMgr.Len())
}

// 删除链接
func (connMgr *ConnManger) Remove(conn ziface.IConnect) {
	//保护共享资源map，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除连接信息
	delete(connMgr.connections, conn.GetConnID())

	fmt.Println("connID =", conn.GetConnID(), "remove from ConnManager successfully : conn num = ", connMgr.Len())
}

// 根据当前connID获取链接
func (connMgr *ConnManger) Get(connID uint32) (ziface.IConnect, error) {
	//保护共享资源map，加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		//找到了
		return conn, nil
	} else {
		return nil, errors.New("connection not FOUND!")
	}
}

// 得到当前连接总数
func (connMgr *ConnManger) Len() int {
	return len(connMgr.connections)
}

// 清除并终止所有连接
func (connMgr *ConnManger) ClearConn() {
	//保护共享资源map，加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	//删除conn并停止conn的工作
	for connID, conn := range connMgr.connections {
		//停止
		conn.Stop()
		//删除
		delete(connMgr.connections, connID)
	}

	fmt.Println("Clear All connections succ! conn")
}
