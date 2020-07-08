package scrapy

import (
	"github.com/Shopify/sarama"
	"os"
	"os/signal"
	"syscall"
	"time"

	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/go-redis/redis"
	"github.com/nsqio/go-nsq"
)

var (
	Stop = make(chan os.Signal, 0)
)

func init() {
	signal.Notify(Stop, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		v := <-Stop
		log.Info("consumer is close.", v)
		os.Exit(0)
	}()
}

func DecodeMessage(b []byte) (*nsq.Message, error) {
	var msg nsq.Message
	msg.Timestamp = time.Now().Unix()
	msg.Body = b
	return &msg, nil
}

type IConsumer interface {
	Init()
	Run()
}

type kafkaConsumer struct {
	Addrs      []string `yaml:"addrs"`
	Topic      string   `yaml:"topic"`
	msgHandler nsq.Handler
	consumer   sarama.Consumer
	Limit      int
	WaitGroupWrap
}

func (k *kafkaConsumer) HandleMessage(message *nsq.Message) error {
	log.Debug(string(message.Body))
	return nil
}

func (k *kafkaConsumer) SetHandler(handler nsq.Handler) IConsumer {
	k.msgHandler = handler
	return k
}

func (k *kafkaConsumer) Init() {
	var err error
	k.consumer, err = sarama.NewConsumer(k.Addrs, nil)
	if err != nil {
		panic("new kafka consumer client error :" + err.Error())
	}
	log.Debug("kafka consumer init ok")
}

func (k *kafkaConsumer) IsAvailable() bool {
	topics, err := k.consumer.Topics()
	if err != nil {
		log.Errorf("ERROR: Unable to list kafka topics, err=[%v]", err)
		return false
	}

	for i, topic := range topics {
		log.Debugf("\tTopic[%d]: [%s]\n", i, topic)
	}
	return true
}

func (k *kafkaConsumer) Run() {
	var ch = make(chan string, 10000)
	if k.msgHandler == nil {
		k.SetHandler(k)
	}
	k.Add(1)
	go func() {
		defer k.Done()
		for {
			select {
			case v := <-ch:
				msg, err := DecodeMessage([]byte(v))
				if err != nil {
					log.Warning(msg)
					continue
				}
				err = k.msgHandler.HandleMessage(msg)
				if err != nil {
					log.Warning(err)
				}
			default:
				time.Sleep(time.Second)
			}
		}
	}()
	partitionList, err := k.consumer.Partitions(k.Topic)
	if err != nil {
		panic(err)
	}
	for partition := range partitionList {
		//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
		//如果该分区消费者已经消费了该信息将会返回error
		//sarama.OffsetNewest:表明了为最新消息
		pc, err := k.consumer.ConsumePartition(k.Topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			log.Error(err)
			continue
		}
		go func(pc sarama.PartitionConsumer) {
			//Messages()该方法返回一个消费消息类型的只读通道，由代理产生
			defer pc.AsyncClose()
			for msg := range pc.Messages() {
				ch <- string(msg.Value)
			}
		}(pc)
	}
	k.Wait()
}

type nsqConsumer struct {
	Urls       []string `json:"urls"`
	Topic      string   `json:"Topic"`
	Channel    string   `json:"channel"`
	consumer   *nsq.Consumer
	msgHandler nsq.Handler
	Limit      int
}

func (n *nsqConsumer) Init() {
	panic("implement me")
}

func (n *nsqConsumer) Run() {
	var waiter WaitGroupWrap
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
			time.Sleep(time.Duration(1000/n.Limit) * time.Millisecond)
			n.consumer.AddHandler(n.msgHandler)
			err = n.consumer.ConnectToNSQDs(n.Urls)
			if nil != err {
				time.Sleep(100 * time.Millisecond)
			}
			select {}
		})
	}
}

func (n *nsqConsumer) SetHandler(f nsq.Handler) IConsumer {
	n.msgHandler = f
	return n
}

func (n *nsqConsumer) HandleMessage(msg *nsq.Message) (err error) {
	return
}

type RedisConsumer struct {
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

func (r *RedisConsumer) SetHandler(handler nsq.Handler) IConsumer {
	r.f = handler
	return r
}

func (r *RedisConsumer) HandleMessage(msg *nsq.Message) (err error) {
	log.Debug(String(msg.Body))
	return
}

func (r *RedisConsumer) Run() {
	if r.f == nil {
		r.SetHandler(r)
	}
	for {
		str, err := r.c.LPop(r.Topic).Result()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		s := String(str)
		if s.Empty() {
			continue
		}
		msg, err := DecodeMessage(s.Decode())
		if err != nil {
			log.Warning(msg)
			continue
		}
		err = r.f.HandleMessage(msg)
		if err != nil {
			log.Error(err)
		}
		time.Sleep(time.Duration(1000/r.Limit) * time.Millisecond)
	}
}
