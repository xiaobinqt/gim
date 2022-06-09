package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

var serverIP string
var serverPort int

func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "服务器ip")
	flag.IntVar(&serverPort, "port", 8888, "服务器端口")
}

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string // 当前客户端名称
	conn       net.Conn
	flag       int // 当前客户模式
}

func NewClient(serverIP string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIP,
		ServerPort: serverPort,
		flag:       9999,
	}

	// 连接 server
	conn, err := net.Dial("tcp", serverIP+":"+strconv.Itoa(serverPort))
	if err != nil {
		fmt.Println("net dial err:", err)
		return nil
	}
	client.conn = conn

	return client
}

func (c *Client) menu() bool {
	var flagMenu int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flagMenu)

	if flagMenu >= 0 && flagMenu <= 3 {
		c.flag = flagMenu
		return true
	} else {
		fmt.Println(">>>> 输入错误，请输入合法数字，请重新输入 <<<<")
		return false
	}
}

func (c *Client) UpdateName() bool {
	fmt.Println("请输入新的用户名：")
	fmt.Scanln(&c.Name)

	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println(">>>>>> 发送消息失败 <<<<<<")
		return false
	}

	return true
}

func (c *Client) PublicChat() {
	// 提示用户输入
	var chatMsg string
	fmt.Println(">>>> 请输入消息：")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println(">>>>>> 发送公聊消息失败 <<<<<<")
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>> 请输入消息或输入 exit 退出：")
		fmt.Scanln(&chatMsg)
	}
}

// 查询在线用户
func (c *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println(">>>>>> SelectUser err <<<<<<", err)
		return
	}
}

func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	c.SelectUser()
	fmt.Println(">>> 请输入聊天对象，或输入 exit 退出：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>> 请输入消息，或输入 exit 退出：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println(">>>>>> 发送私聊消息失败 <<<<<<")
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>> 请输入消息或输入 exit 退出：")
			fmt.Scanln(&chatMsg)
		}

		c.SelectUser()
		fmt.Println(">>> 请输入聊天对象，或输入 exit 退出：")
		fmt.Scanln(&remoteName)
	}
}

func (c *Client) DealResponse() {
	// io.copy 永久阻塞,一旦 c.conn 有数据，就直接 copy 到 stdout
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {

		}

		switch c.flag {
		case 1: // 公聊模式
			//fmt.Println("公聊模式")
			c.PublicChat()
			break
		case 2: // 私聊模式
			//fmt.Println("私聊模式")
			c.PrivateChat()
			break
		case 3: // 更新用户名
			c.UpdateName()
			break
		}
	}
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println(">>>>>> 连接服务器失败 <<<<<<")
		return
	}

	// 单独开启一个 goroutine 来处理服务器返回的消息
	go client.DealResponse()

	fmt.Println(">>>>>> 连接服务器成功 <<<<<<")

	client.Run()
}
