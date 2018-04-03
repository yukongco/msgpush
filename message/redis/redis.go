package libredis

import (
	"fmt"
	"io"

	"github.com/gomodule/redigo/redis"
	"github.com/yukongco/msgpush/common/check"
	"github.com/yukongco/msgpush/message/conf"
)

type RedisPool struct {
	pool *redis.Pool
}

var (
	redisPool *RedisPool
)

// init redis pool
func InitRedis() (*RedisPool, error) {
	pool := &redis.Pool{
		MaxIdle:   conf.Conf.RedisConf.MaxIdle,
		MaxActive: conf.Conf.RedisConf.MaxActive,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(conf.Conf.RedisConf.Proto, conf.Conf.RedisConf.HostPort)
			if err != nil {
				return nil, fmt.Errorf("redis dial err: %v", err)
			}

			if _, err = conn.Do("AUTH", conf.Conf.RedisConf.AUTH); err != nil {
				conn.Close()
				return nil, fmt.Errorf("redis set auth is err: %v", err)
			}
			return conn, nil
		},
	}

	c := pool.Get()

	c.Do("ERR", io.EOF)
	if c.Err() != nil {
		return nil, fmt.Errorf("c do err: %v", c.Err())
	}

	redisPool = &RedisPool{
		pool: pool,
	}

	return redisPool, nil
}

// redis set
func AddSet(key string, value interface{}) error {
	c := redisPool.pool.Get()
	_, err := c.Do(check.SADD, key, value)
	if err != nil {
		return fmt.Errorf("add redis set key=%v value=%v err: %v", key, value, err)
	}

	return nil
}

// get redis set
func GetSets(key string) ([]interface{}, error) {
	c := redisPool.pool.Get()

	smeValues, err := redis.Values(c.Do(check.SMEMBERS, key))
	if err != nil {
		tmpStr := fmt.Sprintf("get set by key=%v err: %v", key, err)
		return nil, fmt.Errorf(tmpStr)
	}

	return smeValues, nil
}

// redis del set by key and member
func DelSetMember(key string, value interface{}, mems ...interface{}) error {
	c := redisPool.pool.Get()

	_, err := c.Do(check.SREM, key, value)
	if err != nil {
		tmpStr := fmt.Sprintf("delete set key=%v value=%v err: %v", key, value, err)
		return fmt.Errorf(tmpStr)
	}

	for _, v := range mems {
		v1 := v
		go c.Do(check.SREM, key, v1)
	}

	return nil
}
