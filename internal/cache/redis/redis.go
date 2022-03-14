package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	redisLib "github.com/go-redis/redis"
)

const (
	defaultPort = 6379
	defaultDb   = 0
)

type Client struct {
	libClient *redisLib.Client
	conf      *Config
	putItem   chan *Item
	lock      sync.RWMutex
}

type Config struct {
	Host       string
	Port       int
	Password   string
	DB         int
	PipeLength int
}

type Item struct {
	Key        string
	Value      interface{}
	Expiration time.Duration
}

func NewConfig() *Config {
	host := os.Getenv("REDIS_HOST")

	port := defaultPort
	portEnv, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err == nil && portEnv > 0 {
		port = portEnv
	}

	password := os.Getenv("REDIS_PASSWORD")

	db := defaultDb
	dbEnv, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err == nil && dbEnv > 0 {
		db = dbEnv
	}

	return &Config{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}

func NewClient(conf *Config) *Client {
	libClient := redisLib.NewClient(&redisLib.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       conf.DB,
	})

	return &Client{
		libClient: libClient,
		conf:      conf,
		putItem:   make(chan *Item),
	}
}

func (client *Client) Connect(ctx context.Context) error {
	_, err := client.libClient.Ping().Result()
	if err != nil {
		return err
	}

	go client.listenForItems(ctx)

	return nil
}

func (client *Client) Put(item *Item) {
	client.putItem <- item
}

func (client *Client) Add(item *Item) error {
	client.lock.Lock()
	defer client.lock.Unlock()

	itemBytes, err := json.Marshal(item.Value)
	if err != nil {
		return err
	}

	return client.libClient.Set(item.Key, string(itemBytes), item.Expiration).Err()
}

func (client *Client) AddMultiple(items []*Item) error {
	pipe := client.libClient.Pipeline()

	for _, item := range items {
		itemBytes, err := json.Marshal(item.Value)
		if err != nil {
			return err
		}

		pipe.Set(item.Key, string(itemBytes), item.Expiration)
	}

	_, err := pipe.Exec()
	return err
}

func (client *Client) Get(key string) ([]byte, error) {
	val, err := client.libClient.Get(key).Result()

	switch {
	// key does not exist
	case err == redisLib.Nil:
		return nil, errors.New(fmt.Sprintf("key %v does not exist", key))
	// some other error
	case err != nil:
		return nil, err
	}

	return []byte(val), nil
}

func (client *Client) GetMultiple(keys ...string) ([]byte, error) {
	var result []byte

	pipe := client.libClient.Pipeline()

	for _, key := range keys {
		pipe.Get(key)
	}

	res, err := pipe.Exec()
	if err != nil {
		return nil, err
	}

	var itemsToReturn [][]byte
	for _, item := range res {
		itemsToReturn = append(itemsToReturn, []byte(item.(*redisLib.StringCmd).Val()))
	}

	itemsByte, err := json.Marshal(itemsToReturn)
	if err != nil {
		return nil, err
	}

	result = itemsByte

	return result, nil
}

func (client *Client) Delete(keys ...string) error {
	return client.libClient.Del(keys...).Err()
}

func (client *Client) DeleteAll() error {
	return client.libClient.FlushAll().Err()
}

func (client *Client) listenForItems(ctx context.Context) {
	for {
		select {
		case item := <-client.putItem:
			client.Add(item)
		case <-ctx.Done():
			return
		}
	}
}
