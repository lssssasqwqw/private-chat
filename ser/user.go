package ser

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"net"
	"server-client/config"
)

// ----------------------------User结构----------------------------
type User struct {
	Name         string //无用
	Account      string
	Addr         string
	C            chan string
	Conn         net.Conn
	Chan_account string //聊天对象
	Is_login     bool   //是否登录
	Session      string
}

var num int = 0
var name_ string = "name"

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	stra := string(num)
	user := &User{
		Name:         name_ + stra,
		Account:      "",
		Addr:         userAddr,
		C:            make(chan string, 6),
		Conn:         conn,
		Chan_account: "",
		Is_login:     false,
	}
	num = num + 1
	//go user.Listenmassage()
	return user
}

// 给所有用户发消息
func (this *User) Listenmassage() {
	for {
		msg := <-this.C
		_, err := this.Conn.Write([]byte(msg + "\n"))
		if err != nil {
			config.Log().Error("")
		}
	}
}

// 给当前用户自己发消息
func (this *User) SendToSlef(msg string) bool {
	for i := 0; i < 6; i++ {
		n, err := this.Conn.Write([]byte(msg + "\n"))
		if err != nil {
			config.Log().Error("SendToSlef:", err)
			return false
		} else {
			log.Printf("已成功发送 %d 字节数据", n)
			break // 跳出循环
		}
	}
	return true
}

type redisUser struct { //redis 数据结构体
	Name    string //无用
	Account string
	Addr    string
	//Conn     net.Conn
	Session  string
	Is_login bool //是否登录
}

// 字典转字符串（json）
func (this *User) StructToJson() string {
	userNoC := redisUser{
		Name:    this.Name,
		Account: this.Account,
		Addr:    this.Addr,
		//Conn:     param.Conn,
		Session:  this.Session,
		Is_login: this.Is_login,
	}
	dataType, err := json.Marshal(userNoC)
	if err != nil {
		config.Log().Error("StructToJson:", err)
	} else {
		dataString := string(dataType)
		return dataString
	}
	return ""

}
func (this *User) CheckRedis(redisconn redis.Conn, key string) (bool, *redisUser) {
	r := &redisUser{}
	s := false
	data := config.RedisGet(redisconn, key, 0)
	if len(data) == 0 {
		fmt.Println("session过期，请重新登录！")
	} else {
		err := json.Unmarshal(data, &r)
		if err != nil {
			config.Log().Error("CheckRedis:", err)
		} else {
			s = true
		}
	}
	return s, r
}

func (this *User) Domassage(Onlinemap map[string]*User, msg string) {
	for _, UserName := range Onlinemap {
		UserName.SendToSlef(msg)
	}
}
