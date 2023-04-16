package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var serverIp string
var serverPort int

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn

	flag int
}

// NewClient 创建一个监听
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	// 建立tcp长链接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIp, client.ServerPort))
	if err != nil {
		fmt.Println("dial err:", err)
		return nil
	}
	client.conn = conn
	return client
}

// Menu 命令菜单
func (client *Client) Menu() bool {
	var code int
	fmt.Println("1: 公聊模式")
	fmt.Println("2: 私聊模式")
	fmt.Println("3: 更新用户名")
	fmt.Println("0: 退出")
	_, err := fmt.Scanln(&code)
	if err != nil {
		return false
	}
	if code >= 0 && code <= 3 {
		client.flag = code
		return true
	}
	return false
}

// 包装更新用户名请求
func (client *Client) updateUserName() bool {
	fmt.Println(">>> 请输入用户名")
	_, err := fmt.Scanln(&client.Name)
	if err != nil {
		return false
	}
	sendMsg := "rename|" + client.Name + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println(">>> conn write err:", err)
		return false
	}
	return true
}

// 查询用户信息
func (client *Client) listUser() bool {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println(">>> conn write err:", err)
		return false
	}
	return true
}

// PrivateChat 私聊模式
func (client *Client) privateChat() {
	var remote, chatMsg string
	client.listUser()
	for remote != "exit" {
		fmt.Println(">>> 请输入发送对象，输入exit退出")
		_, err := fmt.Scanln(&remote)
		if err != nil {
			return
		}
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				// 发送消息到服务端
				sendMsg := "to|" + remote + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println(">>> conn write err:", err)
					break
				}
			}
			// 继续监听下一个发送信息，直到exit
			chatMsg = ""
			fmt.Println(">>> 请输入发送内容，输入exit退出")
			_, err := fmt.Scanln(&chatMsg)
			if err != nil {
				return
			}
		}
		// 内循环退出，继续监听输入内容选择用户，知道exit
		remote = ""
		fmt.Println(">>> 请输入发送对象，输入exit退出")
		_, err = fmt.Scanln(&remote)
		if err != nil {
			return
		}
	}
}

// PublicChat 公聊模式
func (client *Client) publicChat() {
	var chatMsg string
	fmt.Println(">>> 请输入发送内容，输入exit退出")
	_, err := fmt.Scanln(&chatMsg)
	if err != nil {
		return
	}
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println(">>> conn write err:", err)
				break
			}
		}
		// 继续监听下一个发送信息，直到exit
		chatMsg = ""
		fmt.Println(">>> 请输入发送内容，输入exit退出")
		_, err := fmt.Scanln(&chatMsg)
		if err != nil {
			return
		}
	}

}

func (client *Client) Run() {
	for client.flag != 0 {
		// 过滤非法的菜单code
		for !client.Menu() {
		}
		// 根据菜单code进行功能调度
		switch client.flag {
		case 1:
			client.publicChat()
			break
		case 2:
			client.privateChat()
			break
		case 3:
			client.updateUserName()
			break
		}
	}
}

func init() {
	// 初始化配置，通过 -ip设置IP地址
	flag.StringVar(&serverIp, "i", "127.0.0.1", "设置服务器的IP地址(默认127.0.0.1)")
	// 初始化配置，通过 -p设置IP端口
	flag.IntVar(&serverPort, "p", 8888, "设置服务器的端口(默认8888)")
}

func (client *Client) DealResponse() {
	_, err := io.Copy(os.Stdout, client.conn)
	if err != nil {
		return
	}
}

func main() {
	// 命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>> 链接服务器失败")
		return
	}
	go client.DealResponse()
	fmt.Println(">>> 链接服务器成功")
	client.Run()

}
