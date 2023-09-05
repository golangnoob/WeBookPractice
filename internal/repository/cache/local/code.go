package local

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/coocood/freecache"

	"webooktrial/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnknownForCode         = errors.New("我也不知发生什么了，反正是跟 code 有关")
)

type localCodeCache struct {
	client        *freecache.Cache
	expireSeconds int
	cached        map[string]*CachedCode
	lock          *sync.Mutex
}

func NewCodeCache() cache.CodeCache {
	return &localCodeCache{
		client:        freecache.NewCache(100 * 1024 * 1024),
		expireSeconds: 5,
		lock:          &sync.Mutex{},
		cached:        make(map[string]*CachedCode, 100),
	}
}

func (l *localCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := l.generateKey(biz, phone)
	// 在这里加锁，不加锁可能会导致一个用户多次调用 Set 方法
	// 可以将锁加在 key 上：map[key]sync.RWMutex， freecache将锁加在segment上：[segmentCount]sync.Mutex
	l.lock.Lock()
	defer l.lock.Unlock()
	ttl, err := l.client.TTL([]byte(key))
	if err == nil && ttl > 2 {
		return ErrCodeSendTooMany
	}
	if err == nil && ttl == 0 {
		return ErrUnknownForCode
	}
	err = l.client.Set([]byte(key), []byte(code), l.expireSeconds)
	if err != nil {
		return err
	}
	l.cached[key] = &CachedCode{
		code:  code,
		count: 3,
	}
	return nil
}

func (l *localCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	key := l.generateKey(biz, phone)
	l.lock.Lock()
	defer l.lock.Unlock()
	code, err := l.client.Get([]byte(key))
	if err != nil {
		return false, ErrUnknownForCode
	}
	if l.cached[key].count == 0 {
		l.client.Del([]byte(key))
		delete(l.cached, key)
		return false, ErrCodeVerifyTooManyTimes
	}
	if string(code) == inputCode {
		l.client.Del([]byte(key))
		delete(l.cached, key)
		return true, nil
	} else {
		l.cached[key].count--
		return false, nil
	}
}

func (l *localCodeCache) generateKey(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type CachedCode struct {
	code  string
	count int
}
