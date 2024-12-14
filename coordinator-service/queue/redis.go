package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Cloud-Deployments/services/coordinator/job"
	"github.com/redis/go-redis/v9"
	"sync"
)

type RedisDriverOpts struct {
	Host     string
	Port     int
	Password string
	DB       int
	Prefix   string
}

type RedisDriver struct {
	RedisDriverOpts

	mutex  sync.RWMutex
	client *redis.Client
}

func NewRedisDriver(opts RedisDriverOpts) *RedisDriver {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       opts.DB,
	})

	return &RedisDriver{
		RedisDriverOpts: opts,

		mutex:  sync.RWMutex{},
		client: client,
	}
}

// Enqueue implements the Driver interface
func (d *RedisDriver) Enqueue(j *job.Job) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	data, err := json.Marshal(j)
	if err != nil {
		return err
	}

	return d.client.LPush(context.Background(), d.Prefix+"queue", data).Err()
}

// Dequeue implements the Driver interface
func (d *RedisDriver) Dequeue() (*job.Job, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	data, err := d.client.RPop(context.Background(), d.Prefix+"queue").Bytes()
	if err != nil {
		return nil, err
	}

	var j job.Job
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}

	return &j, nil
}

// Close implements the Driver interface
func (d *RedisDriver) Close() error {
	return d.client.Close()
}
