package database

import (
	"context"
	"time"

	"github.com/Kong/kong-cloudtrails-integration/model"
	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v8"
)

type Database struct {
	client *redis.Client
	ctx    context.Context
}

func New(host string, username string, password string, db int) *Database {
	ctx := context.TODO()

	client := redis.NewClient(&redis.Options{
		Addr: host,
		DB:   db,
	})

	if username != "" {
		client.Options().Username = username
	}
	if password != "" {
		client.Options().Password = password
	}

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to Connect to Redis: %s", err.Error())
	}

	return &Database{
		client: client,
		ctx:    ctx,
	}
}

func (db *Database) RequestIdsExist(ids ...string) []int {
	results := make([]int, 0, len(ids))

	cmds, err := db.client.Pipelined(db.ctx, func(pipe redis.Pipeliner) error {
		for _, v := range ids {
			pipe.Exists(db.ctx, v)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, cmd := range cmds {
		results = append(results, int(cmd.(*redis.IntCmd).Val()))

	}

	return results
}

func (db *Database) SetRequestIds(al *model.AuditLogs) {

	pipe := db.client.TxPipeline()
	for k, v := range al.Logs {
		ttl := time.Duration(v.Ttl) * time.Second
		pipe.Set(db.ctx, k, true, ttl)
	}
	_, err := pipe.Exec(db.ctx)
	if err != nil {
		log.Fatalf("Error setting keys to redis: %s", err.Error())
	}

}
