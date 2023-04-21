package config

import (
	"database/sql"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"
)

const (
	Expiretime = 20000 //  redis session保存时间,减少二次登录2
	Db0        = 0     //session 保存的数据库

	Waitsendtime = 3600 * 2
	Db1          = 1 //信息未发生的数据库
)

// ----------------------------表结构----------------------------
// mysql user表结构体
type Person struct {
	Account  *string
	Name     *string
	Password *string
	Temnum   *string
	Friends  *string
	Waitsend string `gorm:"default:null"`
	gorm.Model
	//CreatedAt time.Time `gorm:"column:created_at;default:0000-00-00 00:00:00;not null"`
}

// mysql配置
var sqlConfig = map[string]string{
	"ip":       "127.0.0.1",
	"port":     "3306",
	"password": "Aa112233",
	"dbname":   "data_1",
	"user":     "root",
}

// redis配置
var redisConfig = map[string]string{
	"ip":   "127.0.0.1",
	"port": "6379",
}

// ----------------------------原生mysql连接 数据库连接配置----------------------------
func Conn_sql() *sql.DB {
	db, err := sql.Open("mysql", "root:Aa112233@tcp(127.0.0.1:3306)/data_1?charset=utf8")
	if err != nil {
		fmt.Println("conn is fail ....")
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("连接数据库失败", err)
	}
	return db
}

// --------------------------------gorm连接 数据库配置--------------------------------
func sqlDsn() string {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		sqlConfig["user"], sqlConfig["password"], sqlConfig["ip"], sqlConfig["port"], sqlConfig["dbname"])
	return dsn
}

func SqlConnet() (db *gorm.DB) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Warn, // 日志级别为info
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 彩色打印
		},
	)
	dsn := sqlDsn() // 本地数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})

	if err != nil {
		Log().Error("连接数据库失败1:", err)
		panic(err)
	}
	sqlDB, err := db.DB()
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err != nil {
		Log().Error("连接数据库失败2:", err)
		panic(err)
	}

	db.AutoMigrate(&Person{})
	return db
}

// ---------------------------- Redis连接 数据库连接配置----------------------------

func RedisConn() redis.Conn {
	//通过go向redis写入数据和读取数据
	// 1. 链接到redis
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisConfig["ip"], redisConfig["port"]))
	if err != nil {
		Log().Error("redis.Dial err = ", err)
		panic(err)
	}
	return conn
}
