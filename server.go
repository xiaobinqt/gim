package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// 监听 message, 并发送给所有在线用户
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Broadcast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]%s:%s",
		user.Addr, user.Name, msg)
	s.Message <- sendMsg
}

func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)

	user.Online()

	// 监听用户是否活跃的 channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if n == 0 { // 客户端关闭
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("read error:", err)
				break
			}

			// 提取用户发送的消息,去除 \n
			msg := string(buf[:n-1])

			user.DoMessage(msg)

			// 用户的任意消息，都认为用户是活跃的
			isLive <- true
		}
	}()

	// 阻塞当前 handler
	for {
		select {
		case <-isLive:
		// 客户端活跃
		case <-time.After(1 * time.Hour):
			// 客户端超时,将当前 user 强制关闭

			user.SendMsg("长时间不活跃，你被踢了")

			// 销毁资源
			close(user.C)
			conn.Close()
			return
		}
	}

}

func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", s.Ip+":"+strconv.Itoa(s.Port))
	if err != nil {
		fmt.Println("listen error:", err)
		return
	}
	defer listener.Close()
	// socket accept

	// 启动监听 message
	go s.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener error:", err)
			continue
		}

		// do handle
		go s.Handler(conn)
	}
}
