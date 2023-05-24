package ziface

/*
连接管理模块
*/
type IConnManager interface {
	//添加链接
	Add(conn IConnect)
	//删除链接
	Remove(conn IConnect)
	//根据当前connID获取链接
	Get(connID uint32) (IConnect, error)
	//得到当前连接总数
	Len() int
	//清除并终止所有连接
	ClearConn()
}
