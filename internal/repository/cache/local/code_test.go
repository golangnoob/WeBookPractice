package local

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCode(t *testing.T) {
	cache := NewCodeCache()
	ctx := context.Background()
	fmt.Println("写入本地缓存操作:", cache.Set(ctx, "login",
		"1234567891", "123456"))
	fmt.Println("写入本地缓存操作:", cache.Set(ctx, "login",
		"1234567891", "456789"))
	for i := 0; i < 4; i++ {
		go func() {
			ok, err := cache.Verify(ctx, "login",
				"1234567891", "456789")
			fmt.Println(ok, err)
		}()
	}
	time.Sleep(time.Second * 5)

}
