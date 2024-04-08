package Core

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/kxg3030/shermie-proxy/Contract"
	"github.com/kxg3030/shermie-proxy/Log"
	"github.com/kxg3030/shermie-proxy/Utils"
	"github.com/viki-org/dnscache"
)

var proxyIsSet bool
var lock = &sync.Mutex{}

type HttpRequestEvent func(message []byte, request *http.Request, resolve ResolveHttpRequest, conn net.Conn) bool
type HttpResponseEvent func(message []byte, response *http.Response, resolve ResolveHttpResponse, conn net.Conn) bool

type Socks5ResponseEvent func(message []byte, resolve ResolveSocks5, conn net.Conn) (int, error)
type Socks5RequestEvent func(message []byte, resolve ResolveSocks5, conn net.Conn) (int, error)

type WsRequestEvent func(msgType int, message []byte, resolve ResolveWs, conn net.Conn) error
type WsResponseEvent func(msgType int, message []byte, resolve ResolveWs, conn net.Conn) error

type TcpConnectEvent func(conn net.Conn)
type TcpClosetEvent func(conn net.Conn)
type TcpServerStreamEvent func(message []byte, resolve ResolveTcp, conn net.Conn) (int, error)
type TcpClientStreamEvent func(message []byte, resolve ResolveTcp, conn net.Conn) (int, error)

const (
	MethodGet     = 0x47
	MethodConnect = 0x43
	MethodPost    = 0x50
	MethodPut     = 0x50
	MethodDelete  = 0x44
	MethodOptions = 0x4F
	MethodHead    = 0x48
	SocksFive     = 0x5
)

type ProxyServer struct {
	nagle                  bool
	to                     string
	proxy                  string
	port                   string
	network                string
	listener               *net.TCPListener
	dns                    *dnscache.Resolver
	OnHttpRequestEvent     HttpRequestEvent
	OnHttpResponseEvent    HttpResponseEvent
	OnWsRequestEvent       WsRequestEvent
	OnWsResponseEvent      WsResponseEvent
	OnSocks5ResponseEvent  Socks5ResponseEvent
	OnSocks5RequestEvent   Socks5RequestEvent
	OnTcpConnectEvent      TcpConnectEvent
	OnTcpCloseEvent        TcpClosetEvent
	OnTcpServerStreamEvent TcpServerStreamEvent
	OnTcpClientStreamEvent TcpClientStreamEvent
}

func NewProxyServer(port string, nagle bool, proxy string, to string, network string) *ProxyServer {
	return &ProxyServer{
		port:    port,
		dns:     dnscache.New(time.Minute * 5),
		nagle:   nagle,
		proxy:   proxy,
		to:      to,
		network: network,
	}
}

func (i *ProxyServer) Install() {
	//if runtime.GOOS == "windows" {
	//	err := Utils.InstallCert("cert.crt")
	//	if err != nil {
	//		Log.Log.Println(err.Error())
	//		return
	//	}
	//	Log.Log.Println("已安装系统证书")
	//	err = Utils.SetSystemProxy(fmt.Sprintf("127.0.0.1:%s", i.port))
	//	if err != nil {
	//		Log.Log.Println(err.Error())
	//		return
	//	}
	//	Log.Log.Println("已设置系统代理")
	//	return
	//}
	Log.Log.Println("非windows系统请手动安装证书并设置代理,可以在根目录或访问http://127.0.0.1/tls获取证书文件")
}

func (i *ProxyServer) UnInstall() {
	if runtime.GOOS == "windows" {
		err := Utils.SetSystemProxy("")
		if err != nil {
			Log.Log.Println(err.Error())
			return
		}
		Log.Log.Println("已关闭系统代理")
		return
	}
}

func (i *ProxyServer) beforeStart() {
	lock.Lock()
	defer func() {
		lock.Unlock()
	}()
	if !proxyIsSet {
		i.Logo()
		// i.Install()
		proxyIsSet = true
	}

}

func (i *ProxyServer) Start() error {
	i.beforeStart()
	Log.Log.Println("0.0.0.0:" + i.port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%s", i.port))
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	i.listener = listener
	i.MultiListen()
	select {}
}

func (i *ProxyServer) Stop() error {
	i.UnInstall()
	return nil
}

func (i *ProxyServer) Logo() {
	logo := ` 
                              $$\                                                                                                     $$\                                     
                              $$ |                                                                                                    $$ |                                    
 $$$$$$$\ $$$$$$\   $$$$$$\ $$$$$$\   $$\   $$\  $$$$$$\   $$$$$$\           $$$$$$\   $$$$$$\   $$$$$$\  $$\   $$\ $$\   $$\         $$ |  $$\  $$$$$$\  $$$$$$$\   $$$$$$\  
$$  _____|\____$$\ $$  __$$\\_$$  _|  $$ |  $$ |$$  __$$\ $$  __$$\ $$$$$$\ $$  __$$\ $$  __$$\ $$  __$$\ \$$\ $$  |$$ |  $$ |$$$$$$\ $$ | $$  |$$  __$$\ $$  __$$\ $$  __$$\ 
$$ /      $$$$$$$ |$$ /  $$ | $$ |    $$ |  $$ |$$ |  \__|$$$$$$$$ |\______|$$ /  $$ |$$ |  \__|$$ /  $$ | \$$$$  / $$ |  $$ |\______|$$$$$$  / $$ /  $$ |$$ |  $$ |$$ /  $$ |
$$ |     $$  __$$ |$$ |  $$ | $$ |$$\ $$ |  $$ |$$ |      $$   ____|        $$ |  $$ |$$ |      $$ |  $$ | $$  $$<  $$ |  $$ |        $$  _$$<  $$ |  $$ |$$ |  $$ |$$ |  $$ |
\$$$$$$$\\$$$$$$$ |$$$$$$$  | \$$$$  |\$$$$$$  |$$ |      \$$$$$$$\         $$$$$$$  |$$ |      \$$$$$$  |$$  /\$$\ \$$$$$$$ |        $$ | \$$\ \$$$$$$  |$$ |  $$ |\$$$$$$$ |
 \_______|\_______|$$  ____/   \____/  \______/ \__|       \_______|        $$  ____/ \__|       \______/ \__/  \__| \____$$ |        \__|  \__| \______/ \__|  \__| \____$$ |
                   $$ |                                                     $$ |                                    $$\   $$ |                                      $$\   $$ |
                   $$ |                                                     $$ |                                    \$$$$$$  |                                      \$$$$$$  |
                   \__|                                                     \__|                                     \______/                                        \______/ 
`
	Log.Log.Println(logo)

}

func (i *ProxyServer) MultiListen() {
	for s := 0; s < 5; s++ {
		go func() {
			for {
				conn, err := i.listener.Accept()
				if err != nil {
					if e, ok := err.(net.Error); ok && e.Timeout() {
						Log.Log.Println("接受连接超时：" + err.Error())
						time.Sleep(time.Second / 20)
					} else {
						Log.Log.Println("接受连接失败：" + err.Error())
					}
					continue
				}
				go i.handle(conn)
			}
		}()
	}
}

func (i *ProxyServer) handle(conn net.Conn) {
	var process Contract.IServerProcesser
	defer func() {
		if i.OnTcpCloseEvent != nil {
			i.OnTcpCloseEvent(conn)
		}
		conn.Close()
	}()
	if i.OnTcpConnectEvent != nil {
		i.OnTcpConnectEvent(conn)
	}
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	// https、ws、wss读取到的数据为：CONNECT xx.com:8080 HTTP/1.1
	peek, err := reader.Peek(3)
	if err != nil {
		return
	}
	peer := ConnPeer{server: i, conn: conn, writer: writer, reader: reader}
	switch peek[0] {
	case MethodGet, MethodPost, MethodDelete, MethodOptions, MethodHead, MethodConnect:
		process = &ProxyHttp{ConnPeer: peer}
	case SocksFive:
		process = &ProxySocks5{ConnPeer: peer}
	default:
		process = &ProxyTcp{ConnPeer: peer}
	}
	process.Handle()
}
