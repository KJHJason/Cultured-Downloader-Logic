package cache

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/cockroachdb/pebble"
)

type DbWrapper struct {
	Db *pebble.DB
}

func (db *DbWrapper) Close() error {
	return db.Db.Close()
}

func NewDb(path string) (*DbWrapper, error) {
	os.MkdirAll(path, constants.DEFAULT_PERMS)
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &DbWrapper{Db: db}, nil
}

func handleErr(err error, logMsg string) {
	logger.MainLogger.Fatalf("%s: %s", logMsg, err)
}

func handleCloserErr(closer io.Closer) {
	err := closer.Close()
	if err == nil {
		return
	}
	// Shouldn't happen but log it and exit(1) to avoid memory leaks
	handleErr(err, "Failed to close cache value")
}

func (db *DbWrapper) Get(key string) []byte {
	if CacheDb == nil {
		return nil
	}

	value, closer, err := CacheDb.Db.Get([]byte(key))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil
		}
		handleErr(err, "Failed to get cache value") // will exit the program, hence no need to return
	}
	defer handleCloserErr(closer)
	return value
}

func (db *DbWrapper) Delete(key string) error {
	err := db.Db.Delete([]byte(key), pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (db *DbWrapper) ResetDb() error {
	iter, err := db.Db.NewIter(nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	batch := db.Db.NewBatch()
	for iter.First(); iter.Valid(); iter.Next() {
		err = batch.Delete(iter.Key(), pebble.Sync)
		if err != nil {
			batch.Close()
			return err
		}
	}

	err = db.SetBatch(batch)
	if err != nil {
		batch.Close()
		return err
	}
	return nil
}

func (db *DbWrapper) SetBatch(batch *pebble.Batch) error {
	err := db.Db.Apply(batch, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (db *DbWrapper) Set(key string, value []byte) error {
	err := db.Db.Set([]byte(key), value, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (db *DbWrapper) SetJson(key string, v any) error {
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return db.Set(key, value)
}

func (db *DbWrapper) SetString(key string, value string) error {
	return db.Set(key, []byte(value))
}

func (db *DbWrapper) SetInt64(key string, value int64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return db.Set(key, buf)
}

func (db *DbWrapper) SetInt(key string, value int) error {
	return db.SetInt64(key, int64(value))
}

func (db *DbWrapper) SetTime(key string, value time.Time) error {
	return db.SetInt64(key, value.UTC().Unix())
}

func (db *DbWrapper) GetJson(key string, v any) error {
	err := json.Unmarshal(db.Get(key), v)
	if err != nil {
		return err
	}
	return nil
}

func (db *DbWrapper) GetString(key string) string {
	return string(db.Get(key))
}

func (db *DbWrapper) GetInt64(key string) int64 {
	value := db.Get(key)
	if value == nil {
		return -1
	}
	return int64(binary.LittleEndian.Uint64(value))
}

func (db *DbWrapper) GetInt(key string) int {
	return int(db.GetInt64(key))
}

func (db *DbWrapper) GetTime(key string) time.Time {
	return time.Unix(db.GetInt64(key), 0)
}
