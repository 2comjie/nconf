# nconf

Go 配置中心库：多数据源加载、合并成配置树、按 key 读取、热更新监听。

## 特性

- **多数据源**：本地文件/目录（`file`）、`embed.FS`（`embedfs`），可扩展自定义 Source
- **多格式**：JSON、YAML/YML、XML、Proto、Gob，可注册自定义编解码器
- **配置树合并**：文件路径自动映射为 key 层级（`db/mysql.yaml` → `db.mysql.*`），多文档自动合并
- **热更新**：file 源基于 fsnotify 监听变更（100ms 防抖），支持 `Watch` 观察者回调
- **泛型 API**：`Get[T]` 一键取值、`WatchedValue[T]` 自动跟随更新的原子值
- **并发安全**：快照 + atomic.Pointer 读取路径无锁

## 安装

```bash
go get github.com/2comjie/nconf@latest
```

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	"github.com/2comjie/nconf"
	"github.com/2comjie/nconf/file"
)

func main() {
	c := nconf.New(nconf.WithSource(file.NewSource("config.yaml")))
	if err := c.Load(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// 泛型按键读取（key 用点号分隔层级）
	addr, _ := nconf.Get[string](c, "server.http.addr")
	fmt.Println(addr)

	// 或整体 Scan 到结构体（支持 proto.Message）
	var cfg MyConfig
	_ = c.Scan(&cfg)
}
```

## key 路径映射规则

数据源返回的每个 `Document` 的 `Path` 会被映射为配置树中的层级：

| 文件路径 | 映射后的 key 前缀 |
|---|---|
| `server.yaml` | `server.*` |
| `db/mysql.yaml` | `db.mysql.*` |
| `a.b.yaml` | `a.b.*` |
| 根文档（Path 为空） | 内容直接合并到根，必须是对象 |

同一个 Source 内多个文件映射到同一路径会报错；不同 Source 的同名 key 后者覆盖前者（按 `WithSource` 顺序）。XML 文件的根元素若与文件名一致会自动去掉一层嵌套。

## 数据源

| Source | 用法 | 热更新 |
|---|---|---|
| 本地文件/目录 | `file.NewSource("conf/")` | 支持（fsnotify，递归监听子目录，100ms 防抖） |
| embed.FS | `embedfs.NewSource(fsys, "config.yaml")` | 不支持（返回 `ErrWatchUnsupported`，加载后静态） |

自定义 Source 实现接口即可：

```go
type Source interface {
	Load() ([]Document, error)   // 返回一批配置文档
	Watch() (Watcher, error)     // 不支持时返回 ErrWatchUnsupported
}
```

## 热更新

```go
// 方式一：Watch 某个 key，变化时回调
_ = c.Watch("server.http.addr", func(key string, value nconf.Value) {
	addr, _ := value.String()
	fmt.Println("changed:", addr)
})

// 方式二：WatchedValue[T] 包装成原子值，业务随时 Load() 都是最新的
var conf nconf.WatchedValue[HTTPConfig]
_ = conf.Init(c, "server.http")
current := conf.Load()
```

注意：`Value(key)` 拿到的 `Value` 会随 reload 自动更新，但每次 `Load()` 读到的是当时的快照；观察者回调在独立 goroutine 中异步触发。

## Value API

```go
v := c.Value("database.maxOpenConns")
v.Int()       // 数值互转（int/uint/float 自动转换，string 尝试解析）
v.String()    // 数字/布尔转字符串
v.Bool()
v.Float()
v.Duration()  // 注意：底层是 Int()，即纳秒数；配置里请写整数（如 3000000000 表示 3s）
v.Slice()     // []Value
v.Map()       // map[string]Value
v.Scan(&out)  // 反序列化到结构体 / proto.Message
```

## 扩展编解码器

```go
type MyCodec struct{}

func (MyCodec) Marshal(v any) ([]byte, error)   { ... }
func (MyCodec) Unmarshal(data []byte, v any) error { ... }
func (MyCodec) Name() string                     { return "toml" }

encoding.RegisterCodec(MyCodec{})  // 之后 .toml 文件自动可用
```

## 示例

见 [examples](./examples)：

- [basic](./examples/basic) — 文件加载、`Get[T]`、`Value`、`Scan`
- [watch](./examples/watch) — `Watch` 热更新回调（自包含，自动改写临时文件演示）
- [embed](./examples/embed) — embed.FS 源 + `WatchedValue[T]`

```bash
cd examples/basic && go run .
```

## License

Apache-2.0
