package storecli

import (
	"errors"
	"time"
)

func Set(key string, json string, duration time.Duration) error {
	if key == "" {
		return errors.New("key is empty")
	}
	getCacheStore().set(key, json, duration)
	return nil
}

func Get(key string) (string, error) {
	if key == "" {
		return "", errors.New("key is empty")
	}
	return getCacheStore().get(key), nil
}

func Del(key string) error {
	getCacheStore().delete(key)
	return nil
}

func Ttl(key string) time.Duration {
	if key == "" {
		return -2
	}
	return getCacheStore().ttl(key)
}
