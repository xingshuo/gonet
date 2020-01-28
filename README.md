gonet
=========
    network library based on golang.

Wiki
----
    基于golang标准库封装的网络库,目前只支持tcp模式.


设计思想:
-----
    提供语义更清晰,更简洁的以下3个抽象模块:
	  网络应用层抽象:
	      连接器Dialer 和 监听器Listener.
    
	  网络连接层抽象:
	      网络连接的封装Conn,提供读写缓存区能力,读写流程在不同goroutine中处理 .
    
	  以及业务层抽象:
		  type Receiver interface {
			OnConnected(s Sender) error //连接建立
			OnMessage(s Sender, b []byte) (n int, err error) //接收消息
		  }
		  使用者只需要在创建Dialer或Listener时指定Receiver接口的实例

	下面是模块间关系流程图:
![flowchart](https://github.com/xingshuo/gonet/blob/master/flowchart.png)
    
使用流程:
-----
    Dialer(前端):
      启动:
        //ClientConnReceiver为Receiver实例
        d,err := gonet.NewDialer(serverURL, ClientConnReceiver, dialOptions...)
        //省略异常处理
        d.Start() //注意:这里不会阻塞,只会返回连接成功或失败.
                  //成功后自动开启断线重连机制,并根据dialOptions控制重试相关参数
      退出:
        d.Shutdown()

    Listener(后端):
      启动:
        //ServerConnReceiver为Receiver实例
        l,err := gonet.NewListener(bindURL, ServerConnReceiver)
        //省略异常处理
        l.Serve() //注意:这里会阻塞在socket的Accept循环,为每个新连接启动单独Goroutine处理
      退出:
        l.GracefulStop() //其实目前这个退出并不优雅(不保证所有Accept出的连接网络流缓存都读写完,只是暴力的断开连接),
                         //但保证一旦执行了GracefulStop,则Serve循环一定在其执行完成后才退出

前置环境
-----
    golang

支持平台
-----
    Linux/Windows

安装
-----
    1. git clone https://github.com/xingshuo/gonet.git
    2. 
      Windows:
          cd examples\helloworld && .\build.bat
      Linux:
          cd examples/helloworld && sh build.sh

运行
-----
    Windows:
       DIR:
           examples\helloworld
       CMD:
          .\server.exe
          .\client.exe (可启多个)
    Linux:
       DIR:
           examples/helloworld
       CMD:
          ./server.exe
          ./client.exe (可启多个)

测试
-----
    1.Dialer断线重连机制:
        Ctrl + C 关闭Server端,观察Client端断线重连失败消息.
        失败几次后,重新启动Server端,观察断线重连成功消息以及后续前后端心跳包输出
    2.退出流程(Dialer Shutdown与Listener GracefulStop)
        前后端皆由Ctrl + C触发,观察输出log是否异常