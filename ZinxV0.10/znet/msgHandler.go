package znet

import (
	"fmt"
	"strconv"
	"utils"
	"ziface"
)

/*
消息处理模块的实现
*/

type MsgHandle struct {
	//存放每个MsgID
	Apis map[uint32]ziface.IRouter
	//负责Worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	//业务工作Worker池的worker数量
	WorkPoolSize int32
}

// 提供一个初始化/创建MsgHandle方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:         make(map[uint32]ziface.IRouter),
		WorkPoolSize: int32(utils.GlobalObject.WorkerPoolSize), //从全局变量中获取
		TaskQueue:    make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

// 调度/执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	//找到当前受到请求的request中的message id
	//1 从Request中找到msgID
	handler, ok := mh.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("api msgID = ", request.GetMsgId(), "is NOT FOUND! Need Register!")
	}
	//2 根据MsgID调度对应router业务即可
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		//id已经注册
		panic("repeat api , msgID = " + strconv.Itoa(int(msgID))) //strconv.Itoa把int转为字符串用来拼接
	}

	//2 添加msg与API的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api MsgID = ", msgID, "succ!")
}

// 启动一个Worker工作池（开启工作池的动作只能发生一次，一个Zinx框架只能有一个worker工作池）
func (mh *MsgHandle) StartWorkPool() {
	//根据WorkerPoolSize来分别开启Worker，每个Worker用一个go来承载
	for i := 0; i < int(mh.WorkPoolSize); i++ {
		//一个worker被启动
		//1 当前的WOrker对应的channel消息队列 开辟空间 第0个Worker 就用第0个channel...
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//2 启动当前的Worker，阻塞等待消息从channel传递进来
		go mh.startOneWorker(i, mh.TaskQueue[i])
	}
}

// 启动一个Worker工作流程
func (mh *MsgHandle) startOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, "is starting ...")

	//不断的阻塞等待对应消息队列的消息
	for {
		select {
		//如果有消息过来，出列的就是一个客户端的Request。执行当前的Request所绑定业务
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 将消息交给TaskQueue，由Worker处理
func (mh *MsgHandle) SendMsgTaskQueue(request ziface.IRequest) {
	//1 将消息平均分配给不同的worker
	//根据客户端建立的ConnID来进行分配
	workerID := request.GetConnection().GetConnID() % uint32(mh.WorkPoolSize)
	fmt.Println("Add ConnID =", request.GetConnection().GetConnID(),
		"request MsgID =", request.GetMsgId(),
		"to WorkID = ", workerID)

	//2 将消息发送给对应的workerTaskQueue
	mh.TaskQueue[workerID] <- request
}
