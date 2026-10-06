// embed 演示两个能力：
//  1. embedfs 源：把配置文件编译进二进制（不支持 watch）
//  2. WatchedValue[T]：把某个 key 包装成"永远是最新值"的原子变量，业务侧直接 Load()
package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/2comjie/nconf"
	"github.com/2comjie/nconf/embedfs"
)

//go:embed config.yaml
var confFS embed.FS

type HTTPConfig struct {
	Addr          string `json:"addr"`
	ReadTimeoutMs int64  `json:"readTimeoutMs"`
}

func main() {
	c := nconf.New(nconf.WithSource(embedfs.NewSource(confFS, "config.yaml")))
	if err := c.Load(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// WatchedValue：Init 后内部自动 Watch，业务侧任何时候 Load() 都是最新值
	var httpConf nconf.WatchedValue[HTTPConfig]
	if err := httpConf.Init(c, "server.http"); err != nil {
		log.Fatal(err)
	}

	conf := httpConf.Load()
	fmt.Printf("http config: %+v\n", conf)
}
