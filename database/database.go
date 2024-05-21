package database

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	bolt "go.etcd.io/bbolt"
)

const (
	PERMS = 0600 // rw-------
)

type DbWrapper struct {
	Db *bolt.DB
}

// HandleErr logs the error and exits the program
// Override this function if you want to handle errors differently
var HandleErr = func(err error, logMsg string) {
	logger.MainLogger.Fatalf("%s: %s", logMsg, err)
}

func (db *DbWrapper) Close() error {
	if db.Db == nil {
		return nil
	}
	return db.Db.Close()
}

func NewDb(path string) (*DbWrapper, error) {
	os.MkdirAll(filepath.Dir(path), constants.DEFAULT_PERMS)

	// Options are needed to avoid the program hanging indefinitely when the db is locked
	db, err := bolt.Open(path, PERMS, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, err
	}
	return &DbWrapper{Db: db}, nil
}

func (db *DbWrapper) Delete(bucket, key string) error {
	err := db.Db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return nil
		}

		return bucket.Delete([]byte(key))
	})
	return err
}

func (db *DbWrapper) DeleteKeyValueOnPrefixAndCond(bucket, prefix string, fn func(key, val []byte) bool) error {
	return db.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}

		var toDelete [][]byte
		iterateOnPrefix(b, prefix, func(k, v []byte) error {
			if fn == nil || fn(k, v) {
				toDelete = append(toDelete, k)
			}
			return nil
		})

		for _, key := range toDelete {
			err := b.Delete(key)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *DbWrapper) DeleteKeyValueOnPrefix(bucket, prefix string) error {
	return db.DeleteKeyValueOnPrefixAndCond(bucket, prefix, nil)
}

func (db *DbWrapper) DeleteBucket(bucket string) error {
	return db.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		return tx.DeleteBucket([]byte(bucket))
	})
}

type KeyValue struct {
	Key    []byte `json:"Key"`
	Val    []byte `json:"Val"`
	Bucket string `json:"Bucket"`

	KeyStr string `json:"KeyStr"`
	ValStr string `json:"ValStr"`
}

func (ckv KeyValue) GetKey() string {
	return string(ckv.Key)
}

func (ckv KeyValue) GetVal() string {
	return string(ckv.Val)
}

func iterateOnPrefix(b *bolt.Bucket, prefix string, fn func(k, v []byte) error) error {
	c := b.Cursor()
	prefixBytes := []byte(prefix)
	for k, v := c.Seek(prefixBytes); k != nil && bytes.HasPrefix(k, prefixBytes); k, v = c.Next() {
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (db *DbWrapper) Get(bucket, key string) []byte {
	var value []byte
	err := db.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return nil
		}

		value = bucket.Get([]byte(key))
		return nil
	})
	if err != nil {
		HandleErr(err, "Failed to get cache value") // will exit the program, hence no need to return
	}
	return value
}

func (db *DbWrapper) GetKeyValueOnPrefix(bucket, prefix string) []*KeyValue {
	var cacheKeys []*KeyValue
	err := db.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}

		return iterateOnPrefix(b, prefix, func(k, v []byte) error {
			cacheKeys = append(cacheKeys, &KeyValue{
				Key:    k,
				Val:    v,
				KeyStr: string(k),
				ValStr: string(v),
				Bucket: bucket,
			})
			return nil
		})
	})
	if err != nil {
		HandleErr(err, "Failed to get cache on prefix") // will exit the program, hence no need to return
	}

	return cacheKeys
}

func (db *DbWrapper) GetAllKeyValue(bucket string) []*KeyValue {
	var cacheKeys []*KeyValue
	err := db.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			cacheKeys = append(cacheKeys, &KeyValue{
				Key:    k,
				Val:    v,
				KeyStr: string(k),
				ValStr: string(v),
				Bucket: bucket,
			})
			return nil
		})
	})
	if err != nil {
		HandleErr(err, "Failed to get all cache keys") // will exit the program, hence no need to return
	}

	return cacheKeys
}

func (db *DbWrapper) Set(bucket, key string, value []byte) error {
	return db.Db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		return b.Put([]byte(key), value)
	})
}

func (db *DbWrapper) SetJson(bucket, key string, v any) error {
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return db.Set(bucket, key, value)
}

func (db *DbWrapper) SetString(bucket, key string, value string) error {
	return db.Set(bucket, key, []byte(value))
}

func (db *DbWrapper) SetInt64(bucket, key string, value int64) error {
	return db.Set(bucket, key, ParseInt64(value))
}

func (db *DbWrapper) SetInt(bucket, key string, value int) error {
	return db.Set(bucket, key, ParseInt(value))
}

func (db *DbWrapper) SetTime(bucket, key string, value time.Time) error {
	return db.Set(bucket, key, ParseDateTimeToBytes(value))
}

func (db *DbWrapper) GetJson(bucket, key string, v any) error {
	err := json.Unmarshal(db.Get(bucket, key), v)
	if err != nil {
		return err
	}
	return nil
}

func (db *DbWrapper) GetString(bucket, key string) string {
	return string(db.Get(bucket, key))
}

func (db *DbWrapper) GetInt64(bucket, key string) int64 {
	value := db.Get(bucket, key)
	if value == nil {
		return -1
	}
	return ParseBytesToInt64(value)
}

func (db *DbWrapper) GetInt(bucket, key string) int {
	value := db.Get(bucket, key)
	if value == nil {
		return -1
	}
	return ParseBytesToInt(value)
}

func (db *DbWrapper) GetTime(bucket, key string) time.Time {
	value := db.Get(bucket, key)
	if value == nil {
		return time.Time{}
	}
	return ParseBytesToDateTime(value)
}
