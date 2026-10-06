package main

import (
	"fmt"
	"log"

	"github.com/2comjie/nconf"
	"github.com/2comjie/nconf/file"
)

type AppConfig struct {
	Server struct {
		HTTP struct {
			Addr          string `json:"addr"`
			ReadTimeoutMs int64  `json:"readTimeoutMs"`
		} `json:"http"`
		GRPC struct {
			Addr string `json:"addr"`
		} `json:"grpc"`
	} `json:"server"`
	Database struct {
		DSN          string `json:"dsn"`
		MaxOpenConns int64  `json:"maxOpenConns"`
		MaxIdleConns int64  `json:"maxIdleConns"`
	} `json:"database"`
	Features struct {
		EnableMetrics bool `json:"enableMetrics"`
	} `json:"features"`
	AppName string `json:"appName"`
}

func main() {
	c := nconf.New(nconf.WithSource(file.NewSource("config.yaml")))
	if err := c.Load(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// 方式一：泛型 Get 按键读取，key 用点号分隔层级
	addr, err := nconf.Get[string](c, "server.http.addr")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("http addr:", addr)

	maxOpen, err := nconf.Get[int64](c, "database.maxOpenConns")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("max open conns:", maxOpen)

	metrics, err := nconf.Get[bool](c, "features.enableMetrics")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("metrics enabled:", metrics)

	// 方式二：Value() 拿到 Value 后自行转换
	name, err := c.Value("appName").String()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("app name:", name)

	// 方式三：整体 Scan 到结构体
	var app AppConfig
	if err := c.Scan(&app); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("scan result: %+v\n", app)
}
