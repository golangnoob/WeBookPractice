//go:build !k8s

// 没有 k8s 这个编译标签
package config

var Config = config{
	DB: DBConfig{
		DSN: "localhots:13316",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
