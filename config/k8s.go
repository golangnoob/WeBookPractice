//go:build k8s

// Package config 使用 k8s 这个编译标签
package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(webooktrial-mysql:3308)/webook",
	},
	Redis: RedisConfig{
		Addr: "webooktrial-redis:6380",
	},
}
