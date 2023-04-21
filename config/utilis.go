package config

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	//"server-client/ser"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func Log() *logrus.Logger {
	log := logrus.New()
	// 设置输出格式为 text
	log.SetFormatter(&logrus.TextFormatter{})
	// 创建带有颜色的输出
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	return log
}

// 字符串转字节
func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// 字节转字符串
func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// 中文字符串根据索引取值
func StrinfIndex(str string, index int) (string, string) {
	byte := []rune(str)
	length := len(byte)

	if index < 0 {
		if length+index < 0 {
			return "", "字符串超出索引"
		}
		return string(byte[length+index]), ""
	} else {
		if index > length-1 {
			return "", "字符串超出索引"
		}
		return string(byte[index]), ""
	}
}

// 字典转字符串（json）
func MapToJson(param map[string]string) string {
	dataType, err := json.Marshal(param)
	if err != nil {
		Log().Error("MapToJson:", err)
	}
	dataString := string(dataType)
	return dataString
}

// 字符串（json）转字典
func JsonToMap(str string) (bool, map[string]string) {
	var tempMap map[string]string
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		Log().Error("JsonToMap:", err)
		return false, map[string]string{}
	}
	return true, tempMap

}

// 获取格式化时间字符串
func GetNowTime() string {
	time_now := time.Now().Format("2006-01-02 15:04:05")
	return "【" + time_now + "】"
}

// 解密获取msg的int字符串
func Decrypt(str string) string {
	var ll []string
	ll = strings.Split(str, "|")
	Result := ""
	for _, s := range ll {
		d, err := strconv.Atoi(s)
		if err != nil {
			Log().Error("Decrypt:", err)
			return "{}"
		} else {
			AcillCode := (d-3)/2 - 5
			var r rune = rune(AcillCode)
			// 真正可以输出字符
			var ResultStr string = string(r)
			Result += ResultStr
		}
	}
	return Result
}

func RedisSet(redisconn redis.Conn, key string, values interface{}, expirationTime int, num int) {
	// 2、 通过go向redis写入数据 string [key - val]
	_, err := redisconn.Do("SELECT", num)

	if err != nil {
		fmt.Println("redis select failed:", err)
		return
	}
	_, err = redisconn.Do("Set", key, values)
	if err != nil {
		Log().Error("RedisSet1:", err)
	}
	_, err = redisconn.Do("EXPIRE", key, expirationTime)
	if err != nil {
		Log().Error("RedisSet2:", err)

	}
}

func RedisGet(redisconn redis.Conn, key string, num int) []byte {
	_, err := redisconn.Do("SELECT", num)

	if err != nil {
		Log().Warning("redisconn:", err)
		return []byte{}
	}

	data, err := redis.Bytes(redisconn.Do("get", key))
	if err != nil {
		if err == redis.ErrNil {
			Log().Warning("redisconn:", err)
		} else {
			Log().Error("RedisGet1:", err)
		}
		return []byte{}
	} else {
		return data
	}
}

//func main() {
//	param := map[string]string{
//		"code":      "1",
//		"user":      "lzs",
//		"password":  "l",
//		"objective": "登录",
//		"message":   "aaaa",
//	}
//	result := MapToJson(param)
//	fmt.Println(reflect.TypeOf(JsonToMap(result)))
//}
