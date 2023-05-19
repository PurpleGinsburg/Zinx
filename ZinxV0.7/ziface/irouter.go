package ziface

//router路由就是一个指令和对应的处理方式，tcp可以把不同消息对应不同路由方式

/*
    路由抽象接口
	路由里的数据都是IRequest
*/

type IRouter interface {
	//处理conn业务之前的钩子方法hook
	PreHandle(request IRequest)
	//在处理conn业务的主方法hook
	Handle(request IRequest)
	//在处理conn业务之后的钩子方法hook
	PostHandle(request IRequest)
}
