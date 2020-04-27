package crawler

import (
	"encoding/binary"
	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/go-redis/redis"
	"github.com/nsqio/go-nsq"
	"go-scrapy/internal"
	"time"
)

func DecodeMessage(b []byte) (*nsq.Message, error) {
	var msg nsq.Message
	msg.Timestamp = int64(binary.BigEndian.Uint64(b[:8]))
	msg.Attempts = binary.BigEndian.Uint16(b[8:10])
	msg.Body = b
	return &msg, nil
}


type ConsumerInterfaceI interface {
	Init()
	Run()
}

type nsqConsumer struct {
	Urls       []string `json:"urls"`
	Topic      string   `json:"topic"`
	Channel    string   `json:"channel"`
	consumer   *nsq.Consumer
	msgHandler nsq.Handler
	Limit      int
}

func (n *nsqConsumer) Init() {
	panic("implement me")
}

func (n *nsqConsumer) Run() {
	var waiter internal.WaitGroupWrap
	for s := 0; s < len(n.Urls); s++ {
		var err error
		waiter.Wrap(func() {
			defer waiter.Done()
			cfg := nsq.NewConfig()
			cfg.MaxInFlight = 1
			n.consumer, err = nsq.NewConsumer(n.Topic, n.Channel, cfg)
			if nil != err {
				return
			}
			if n.msgHandler == nil {
				n.msgHandler = new(nsqConsumer)
			}
			time.Sleep(time.Duration(1000 / n.Limit) * time.Millisecond)
			n.consumer.AddHandler(n.msgHandler)
			err = n.consumer.ConnectToNSQDs(n.Urls)
			if nil != err {
				time.Sleep(100 * time.Millisecond)
			}
			select {}
		})
	}
	waiter.Wait()
}

func (n *nsqConsumer) SetHandler(f nsq.Handler) ConsumerInterfaceI {
	n.msgHandler = f
	return n
}

func (n *nsqConsumer) HandleMessage(msg *nsq.Message) (err error) {
	return
}

type RedisConsumer struct {
	ch       chan internal.String
	Host     string
	Password string
	Db       int
	Topic    string
	f        nsq.Handler
	c        *redis.Client
	Limit    int
}

func (r *RedisConsumer) Init() {
	opt := go_utils.NewRedisConf(r.Host, r.Password, r.Db)
	r.c = go_utils.NewRedis(opt)
	log.Info("redis consumer inited.")
}

func (r *RedisConsumer) SetHandler(handler nsq.Handler) ConsumerInterfaceI {
	r.f = handler
	return r
}

func (r *RedisConsumer) HandleMessage(msg *nsq.Message) (err error) {
	return
}


func (r *RedisConsumer) Run() {
	if r.f.HandleMessage == nil{
		r.SetHandler(r)
	}
	for {
		str, err := r.c.LPop(r.Topic).Result()
		if err != nil{
			time.Sleep(time.Second)
			continue
		}
		s := internal.String(str)
		if s.Empty(){
			continue
		}
		msg, err := DecodeMessage(s.Decode())
		if err != nil{
			log.Warning(msg)
			continue
		}
		err = r.f.HandleMessage(msg)
		if err != nil{
			log.Error(err)
		}
		time.Sleep(time.Duration(1000 / r.Limit) * time.Millisecond)
	}
}
