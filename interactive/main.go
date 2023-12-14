package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViperV1()
	// 依赖注入
	app := InitApp()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.server.Serve()
	log.Println(err)
}
func initViperV1() {
	cfile := pflag.String("config",
		"config/config.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	// 实时监听配置变更
	viper.WatchConfig()
	// 只能告诉你文件变了，不能告诉你，文件的哪些内容变了
	viper.OnConfigChange(func(in fsnotify.Event) {
		// 比较好的设计，它会在 in 里面告诉你变更前的数据，和变更后的数据
		// 更好的设计是，它会直接告诉你差异。
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("db.dsn"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
