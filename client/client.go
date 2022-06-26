package main

import (
	"flag"
	"fmt"
	"github.com/panjf2000/gnet"
	"os"
)

type clientNg struct {
	*gnet.EventServer
	ChanelMsg chan string
	Conn      gnet.Conn
	Target    gnet.Conn
	Name      string
}

type Client struct {
	Intranet  *clientNg
	Extranet  *clientNg
	ChanelMsg chan []byte
	ExitMsg   chan bool
}

var global_client *Client

func initClientNg(host, port string) *clientNg {
	cg := new(clientNg)
	cg.ChanelMsg = make(chan string)
	client, err := gnet.NewClient(cg)
	ck, err := client.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		panic(err.Error())
	}
	_ = client.Start()
	cg.Conn = ck
	if host != "" {
		fmt.Println("远程服务器连接成功\n")
	} else {
		fmt.Println("本地服务器连接成功\n")
	}

	return cg
}

func initClient(localHost, remoteHost, inPort, exPort string) *Client {
	client := new(Client)
	client.ChanelMsg = make(chan []byte)
	client.Intranet = initClientNg(localHost, inPort)
	client.Extranet = initClientNg(remoteHost, exPort)
	client.Intranet.Name = "内网客户端"
	client.Extranet.Name = "外网客户端"
	client.Intranet.Target = client.Extranet.Conn
	client.Extranet.Target = client.Intranet.Conn
	return client
}

func (cg *clientNg) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	err := cg.Target.AsyncWrite(packet)
	if err != nil {
		fmt.Println(cg.Name, "数据发送错误", err.Error())
		_ = c.SendTo([]byte("服务端转发异常！"))
	}
	cg.Target.Wake()
	return
}

func main() {
	var port string
	var remote string
	flag.StringVar(&port, "p", "9501", "需要穿透的端口")
	flag.StringVar(&remote, "h", "0.0.0.0", "远程服务地址")
	flag.Parse()
	global_client = initClient("", remote, port, "9000")
	fmt.Println("[服务启动成功][ok]|本地映射端口:", port, "｜外网访问地址：", remote, ":9001\n")
	select {
	case <-global_client.ExitMsg:
		fmt.Println("程序退出")
		os.Exit(0)
	}
}
