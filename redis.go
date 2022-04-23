package redislog

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"sync"
)

// Record Represent a set of a Request passing from the client
type Record struct {
	RemoteAddr    string
	URL           string
	AccessTime    int64
	TimeExecuted  int64
	BodyBytesSent int64
}

const StreamKey = "api-request-log"
const RecordIDsKey = "api-request-record-ids"

var client *redis.Client
var once sync.Once

// Client return a redis client
func Client() *redis.Client {
	once.Do(func() {
		// Maybe you should instantiate redis client by reading config file
		client = redis.NewClient(&redis.Options{
			Network: "tcp",
			Addr:    "localhost:6379",
			DB:      0,
		})
	})
	return client
}

// SendRecord send a record to redis server, two things are done as follows:
// 1. add a entry to the stream
// 2. push a entry id to the list
func SendRecord(record Record) {
	// add a entry by call .XAdd method
	xaddCmd := Client().XAdd(context.Background(), &redis.XAddArgs{
		Stream: StreamKey,
		ID:     "*",
		Values: map[string]interface{}{
			"remote_addr":     record.RemoteAddr,
			"url":             record.URL,
			"access_time":     record.AccessTime,
			"time_executed":   record.TimeExecuted,
			"body_bytes_sent": record.BodyBytesSent,
		},
	})
	if xaddCmd.Err() != nil {
		panic(xaddCmd.Err())
	}
	recordID := xaddCmd.Val()
	// push the id to the list by call .LPush method
	lpushCmd := Client().LPush(context.Background(), RecordIDsKey, recordID)
	if lpushCmd.Err() != nil {
		panic(lpushCmd.Err())
	}
}

// ReadRecord read a record from redis, three things are done as follows:
// 1. retrieve a entry id from the list
// 2. retrieve a entry from the stream via the entry id
// 3. after retrieving the entry, delete the entry from the stream
func ReadRecord() (Record, bool) {
	// retrieve record id from the redis list
	lpopCmd := Client().LPop(context.Background(), RecordIDsKey)
	recordID := lpopCmd.Val()
	if recordID == "" {
		return Record{}, false
	}
	// read the record from the stream
	xreadCmd := Client().XRead(context.Background(), &redis.XReadArgs{
		Streams: []string{StreamKey, recordID},
		Count:   1,
		Block:   0,
	})
	if xreadCmd.Err() != nil {
		panic(xreadCmd.Err())
	}
	// if read successfully, we should remove record from the stream
	xdelCmd := Client().XDel(context.Background(), StreamKey, recordID)
	if xdelCmd.Err() != nil {
		panic(xdelCmd.Err())
	}
	record := xreadCmd.Val()[0].Messages[0].Values

	accessTime, _ := strconv.ParseInt(record["access_time"].(string), 10, 64)
	timeExecuted, _ := strconv.ParseInt(record["time_executed"].(string), 10, 64)
	bodyBytesSent, _ := strconv.ParseInt(record["body_bytes_sent"].(string), 10, 64)

	return Record{
		RemoteAddr:    record["remote_addr"].(string),
		URL:           record["url"].(string),
		AccessTime:    accessTime,
		TimeExecuted:  timeExecuted,
		BodyBytesSent: bodyBytesSent,
	}, true
}
