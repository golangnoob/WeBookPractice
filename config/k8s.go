//go:build k8s

// 使用 k8s 这个编译标签
package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(webooktrial-mysql:11309)/webook",
	},
	Redis: RedisConfig{
		Addr: "webooktrial-redis:11479",
	},
}
