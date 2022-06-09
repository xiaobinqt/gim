package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string // 当前客户端地址
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播用户已上线
	u.server.Broadcast(u, "已上线")
}

func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()
	u.server.Broadcast(u, "已下线")
}

// 给当前 user 对应的客户端发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg + "\n"))
}

func (u *User) DoMessage(msg string) {
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := fmt.Sprintf("%s:%s",
				user.Name, user.Addr)
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 修改用户名 `rename|newName`
		newName := strings.Split(msg, "|")[1]

		// 判断 name 是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("当前用户名被使用\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("您已修改用户名" + newName + "成功\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("消息格式错误\n,请使用 to|userName|msg 格式\n")
			return
		}

		// 判断 name 是否存在
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("用户不存在\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("消息不能为空\n")
			return
		}

		remoteUser.SendMsg(fmt.Sprintf("%s 对您说 %s", u.Name, content))
	} else {
		u.server.Broadcast(u, msg)
	}
}

func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("发送消息失败...", err.Error())
		}
	}
}
