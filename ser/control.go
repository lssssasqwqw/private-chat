package ser

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"server-client/config"
	"strings"
)

func (this *Server) Domassage() {
	for _, UserName := range this.Onlinemap {
		msg_user := "[" + UserName.Addr + "]" + UserName.Account + "在线。。。\n"
		UserName.SendToSlef(msg_user)
		fmt.Println(msg_user)
	}
}

func (this *Server) Sendwaitmsg() {
	for {
		if len(this.waitsendmsg) > 0 {
			fmt.Println("len(this.waitsendmsg", len(this.waitsendmsg))
			for accepter, msg := range this.waitsendmsg {
				accepterUser, ok := this.Onlinemap[accepter]
				if ok { //接收端是否登录
					if accepterUser.Is_login {
						accepterUser.SendToSlef(msg)
					}
				}
			}
		}
	}
}

// 监听·message 广播Channel 一旦有消息，发送全部用户
func (this *Server) ListenMsg() {
	for {
		msg := <-this.message
		this.Rwmutex.Lock()
		for _, cli := range this.Onlinemap {
			cli.C <- msg
			fmt.Println(cli.Name)
		}
		this.Rwmutex.Unlock()
	}
}

// 广播消息
func (this *Server) Bro(user *User, msg map[string]string, session string) {

	if user.Is_login { //发送端是否登录
		accepterUser, ok := this.Onlinemap[msg["accepter"]]
		if ok { //接收端是否登录
			if accepterUser.Is_login {
				accepterUser.SendToSlef(config.MapToJson(msg))
			}
		} else {
			key := uuid.NewV4().String() + msg["accepter"]
			config.RedisSet(this.redisconn, key, config.MapToJson(msg), config.Waitsendtime, config.Db1)
			var person config.Person

			this.mysqlconn.Debug().Model(&config.Person{}).Where("account = ?", msg["accepter"]).Find(&person)
			fmt.Println("person.Waitsend", person.Waitsend)
			if person.Waitsend == "" {
				person.Waitsend = key
				//this.sqlconn.Debug().Model(&person).Update("waitsend", &key)
				if err := this.mysqlconn.Debug().Save(&person); err != nil {
					fmt.Println("err", err)
				}
			} else {
				person.Waitsend = person.Waitsend + "::" + key
				//this.sqlconn.Debug().Model(&person).Update("waitsend", &key)
				if err := this.mysqlconn.Debug().Save(&person); err != nil {
					fmt.Println("err", err)
				}
			}
		}

	} else { //user 账号没有登录登录
		s, r := user.CheckRedis(this.redisconn, session) //检测redis 里面有没有记录
		if s {                                           //redis 里面有记录,session没有过期
			user.Account = r.Account
			user.Is_login = r.Is_login
			user.Name = r.Name
			user.Session = r.Session
			this.Onlinemap[user.Account] = user
			this.Bro(user, msg, session)
		} else { //redis 里面没有记录,session过期
			fmt.Println("发送端请登录...")
		}
	}
}

// 消息发送
func (this *Server) SendMsg(user *User, map_msg map[string]string) {
	//user.Chan_account = map_msg["accepter"]
	backMap := map[string]string{
		"code":     "1",
		"mode":     "1001",
		"sender":   user.Account,
		"accepter": map_msg["accepter"],
		"content":  map_msg["content"],
		"state":    "",
		"time":     config.GetNowTime(),
	}
	if map_msg["content"] == "11" {
		this.showOnlinemap()
	} else if map_msg["content"] == "22" {
		user.showusermsg()
	}
	this.Bro(user, backMap, map_msg["session"])
}

func (this *User) showusermsg() {
	fmt.Println("showusermsg", this.C)
}

func (this *Server) showOnlinemap() {
	for s, user := range this.Onlinemap {
		fmt.Println("showOnlinemap:", s, user)
	}
}

// 登录
func (this *Server) Login(user *User, map_msg map[string]string) {
	//检查redis数据库Session

	//检查mysql数据库有没有账号信息
	var person []config.Person
	this.mysqlconn.Select("Password", "Waitsend", "id").Where("account = ?", map_msg["account"]).First(&person)
	if len(person) == 0 {
		backMap := map[string]string{
			"code": "-1",
			"mode": "1002",
			"msg":  "账号未注册",
		}
		user.SendToSlef(config.MapToJson(backMap))
		//账号未注册

	} else {
		if *person[0].Password == map_msg["password"] {
			//密码正确...
			//账号不能重复登录
			uuid := uuid.NewV4().String()
			isAccount := false
			for _, user_ := range this.Onlinemap {
				if user_.Account == map_msg["account"] { //二次登录
					user_.Is_login = false //账号挤掉
					user.Account = map_msg["account"]
					user.Name = map_msg["user"]
					user.Session = uuid
					user.Is_login = true
					this.Onlinemap[user.Account] = user

					backMap := map[string]string{
						"code":    "1",
						"mode":    "1002",
						"name":    map_msg["account"],
						"seeeion": uuid,
					}
					user.SendToSlef(config.MapToJson(backMap))
					isAccount = true
				}
			}

			if !isAccount { //第一次登录 走这个条件
				this.Onlinemap[map_msg["account"]] = user
				user.Account = map_msg["account"]
				user.Name = map_msg["user"]
				user.Is_login = true
				user.Session = uuid
				backMap := map[string]string{
					"code":    "1",
					"mode":    "1002",
					"name":    map_msg["account"],
					"seeeion": uuid,
				}
				user.SendToSlef(config.MapToJson(backMap))
				fmt.Println(backMap)
			}
			if person[0].Waitsend != "" {
				waitsend := strings.Split(person[0].Waitsend, "::")
				for _, s := range waitsend {
					msg := config.RedisGet(this.redisconn, s, 1)
					user.SendToSlef(string(msg))
				}
				this.mysqlconn.Debug().Model(&config.Person{}).Where("id = ?", person[0].ID).Update("Waitsend", "")
			}
			userstring := user.StructToJson() //user转为string 写入redis
			config.RedisSet(this.redisconn, uuid, userstring, config.Expiretime, config.Db0)
		} else {
			backMap := map[string]string{
				"code": "-1",
				"mode": "1002",
				"msg":  "密码错误",
			}
			user.SendToSlef(config.MapToJson(backMap))
			//密码错误...
		}
	}
}
