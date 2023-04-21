package ser

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"gorm.io/gorm"
	"io"
	"net"
	"time"

	"server-client/config"
	"sync"
	"unicode/utf8"
	"unsafe"
)
import _ "github.com/go-sql-driver/mysql"

// 定义类
type Server struct {
	IP          string
	Port        int
	Onlinemap   map[string]*User //在线用户
	Rwmutex     sync.RWMutex     //线程锁
	message     chan string      //广播消息
	mysqlconn   *gorm.DB
	redisconn   redis.Conn
	nowtime     string
	waitsendmsg map[string]string //接收端没有在线 需要保存在这个map里面
}

func SubStrDecodeRuneInString(s string, length int) string {
	var size, n int
	for i := 0; i < length && n < len(s); i++ {
		_, size = utf8.DecodeRuneInString(s[n:])
		n += size
	}
	return s[:n]
}

// 构造方法
func Newserver(iP string, port int) *Server {
	server := &Server{
		IP:          iP,
		Port:        port,
		Onlinemap:   make(map[string]*User),
		message:     make(chan string),
		mysqlconn:   config.SqlConnet(),
		redisconn:   config.RedisConn(),
		nowtime:     time.Now().Format("2006-01-02 15:04:05"),
		waitsendmsg: make(map[string]string),
	}
	return server
}

// 类里面的方法
func (this *Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	} else {
		fmt.Println(config.GetNowTime() + "服务启动成功...")
	}

	defer listen.Close()
	go this.ListenMsg()
	//go this.Sendwaitmsg()
	//go this.Bro()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen Accept:", err)
			continue
		}
		go this.head(conn)
	}
}

// 类里面的方法，端口监听到有连接
func (this *Server) head(conn net.Conn) {

	user := NewUser(conn)             //实例化对象 有net对象连接到服务器
	userstring := user.StructToJson() //user转为string 写入redis
	config.RedisSet(this.redisconn, "登录", string(userstring), 200, config.Db0)
	this.Rwmutex.Lock()
	this.Onlinemap[user.Account] = user
	this.Rwmutex.Unlock()

	go func() {
		buf := make([]byte, 512*16)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				config.Log().Error("连接失败:", err)
				user.Is_login = false
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn Read err:", err)
			}
			rbyf_pn := buf[0:n]
			msg := *(*string)(unsafe.Pointer(&rbyf_pn))
			if msg == "\n" {
				continue
			}
			laststr, ok := config.StrinfIndex(msg, -1) //去除末尾为“/n”字符串
			if ok == "" {
				if laststr == "\n" {
					msg = string([]rune(msg)[:len([]rune(msg))-1])
				}
			} else {
				panic(ok)
			}

			//根据map_msg解析数据
			yes, map_msg := config.JsonToMap(config.Decrypt(msg))
			if !yes {
				backMap := map[string]string{
					"code": "-1",
					"mode": "1002",
					"msg":  "输入格式错误！",
				}
				user.SendToSlef(config.MapToJson(backMap))
				continue
			}
			if len(map_msg) != 0 {
				if map_msg["code"] != "1" {
					backMap := map[string]string{
						"code": "-1",
						"mode": "1002",
						"msg":  "code码错误！",
					}
					user.SendToSlef(config.MapToJson(backMap))
					continue
				}

				if map_msg["mode"] == "1001" { //发送消息
					this.SendMsg(user, map_msg)
				} else if map_msg["mode"] == "1002" { //登录
					this.Login(user, map_msg)
				} else if map_msg["mode"] == "1003" {
					this.Domassage()
				} else if map_msg["mode"] == "1004" {
					user.SendToSlef(map_msg["content"])
				}
			}
		}
	}()
	select {}
}
