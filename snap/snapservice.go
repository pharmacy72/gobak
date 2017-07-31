package snap

import (
	"fmt"
	"github.com/arteev/tern"
	"github.com/pharmacy72/gobak/config"
	"gopkg.in/redis.v5"
	"strconv"
	"sync"
	"time"

	"container/list"
	"log"
)

var (
	snapchannel chan Snapper
	done        chan struct{}
	once        sync.Once
	started     bool
	mu          sync.Mutex
	mr          sync.Mutex
	queue       = list.New()

	clientRedis *redis.Client

	Timeout    = time.Second * 1 // Timeout queue check
	CountQueue = 100
)

const (
	CounterStartBackup   = "backupstart"
	CounterErrorBackup   = "backuperror"
	CounterSuccessBackup = "backupsuccess"
	CounterStart         = "start"
	CounterStop          = "stop"
	CounterCheck         = "check"
	CountStats           = "stats"
)

func SendItem(si Snapper) {
	if started {
		snapchannel <- si
	}

}

func init() {
	once.Do(func() {
		snapchannel = make(chan Snapper, 20)
		done = make(chan struct{})
	})
}

func dosend() {
	mu.Lock()
	defer func() {
		mu.Unlock()
		checkbroken()
	}()
	cnt := 0
	for el := queue.Front(); el != nil; el = queue.Front() {
		if sn, ok := el.Value.(Snapper); ok {
			err := sn.Send()
			if err != nil {
				log.Println(err)
				return
			}
			cnt++
			queue.Remove(el)
		} else {
			panic("Error cast element of list to Snap")
		}
	}
	fmt.Println("Count send:", cnt)
}

func checkbroken() {
	mu.Lock()
	defer mu.Unlock()
	if !started {
		return
	}
	for queue.Len() > CountQueue {
		queue.Remove(queue.Front())
	}
}

func newsnap() *SnapItem {
	return &SnapItem{}
}

func Ping(namedb string) {

	if !started {
		return
	}

	s := newsnap()
	s.action = func() error {
		mr.Lock()
		defer mr.Unlock()
		err := clientRedis.SAdd("gobak:dbs", namedb).Err()
		if err != nil {
			return err
		}
		err = clientRedis.HSet("gobak:db:"+namedb, "ping", strconv.FormatInt(time.Now().UTC().Unix(), 10)).Err()
		if err != nil {
			return err
		}
		return nil
	}
	SendItem(s)
}

func BackupDone(namedb, level, size, errvalue string) {

	if !started {
		return
	}
	s := newsnap()
	s.action = func() error {
		mr.Lock()
		defer mr.Unlock()
		err := clientRedis.HMSet("gobak:db:"+namedb+":lastbackup",
			map[string]string{
				"time":  strconv.FormatInt(time.Now().UTC().Unix(), 10),
				"level": level,
				"size":  size,
				"error": errvalue,
			}).Err()
		if err != nil {
			return err
		}
		return nil
	}
	SendItem(s)
}

func CheckDB(namedb, log string, iserror bool) {
	if !started {
		return
	}
	s := newsnap()
	s.action = func() error {
		mr.Lock()
		defer mr.Unlock()
		err := clientRedis.HMSet("gobak:db:"+namedb+":lastcheck",
			map[string]string{
				"time":    strconv.FormatInt(time.Now().UTC().Unix(), 10),
				"log":     log,
				"iserror": tern.Op(iserror, "true", "false").(string),
			}).Err()
		if err != nil {
			return err
		}
		return nil
	}
	SendItem(s)
}

func Stats(namedb, info string, iserror bool) {

	if !started {
		return
	}
	s := newsnap()
	s.action = func() error {
		mr.Lock()
		defer mr.Unlock()
		err := clientRedis.HMSet("gobak:db:"+namedb+":stats",
			map[string]string{
				"time":    strconv.FormatInt(time.Now().UTC().Unix(), 10),
				"info":    info,
				"iserror": tern.Op(iserror, "true", "false").(string),
			}).Err()

		if err != nil {
			return err
		}
		return nil
	}
	SendItem(s)
}

func Incr(namedb, key, field string, delta int64) {
	if !started {
		return
	}
	s := newsnap()
	s.action = func() error {
		mr.Lock()
		defer mr.Unlock()
		err := clientRedis.HIncrBy("gobak:db:"+namedb+":"+key, field, delta).Err()
		if err != nil {
			return err
		}
		return nil
	}
	SendItem(s)
}

func Start() {
	mu.Lock()
	defer mu.Unlock()
	cfg := config.Current()
	if started || !cfg.Redis.Enabled {
		return
	}
	started = true
	if cfg.Redis.Timeout > 0 {
		Timeout = time.Duration(cfg.Redis.Timeout) * time.Millisecond
	}
	CountQueue = cfg.Redis.Queue
	clientRedis = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	pong, err := clientRedis.Ping().Result()
	fmt.Println(pong, err)

	go func() {
		for {
			select {
			case <-done:
				dosend()
				return
			case si := <-snapchannel:
				mu.Lock()
				queue.PushBack(si)
				mu.Unlock()
			}
		}
	}()

	go func() {
		for {
			<-time.After(Timeout)
			dosend()
			if !started {
				break
			}
		}
	}()
}
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if !started {
		return
	}
	started = false
	close(done)
}
