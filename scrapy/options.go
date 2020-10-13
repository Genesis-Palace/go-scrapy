package scrapy

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/Genesis-Palace/requests"
	"github.com/Shopify/sarama"
	"github.com/go-redis/redis"
	"github.com/goinggo/mapstructure"
	"gopkg.in/yaml.v2"
)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

type IBroker interface {
	Init()
	Add(item IItem) bool
}

type Broker struct {
	NsqBroker   *NsqBroker   `yaml:"nsq" json:"nsq_broker"`
	RedisBroker *RedisBroker `yaml:"redis" json:"redis_broker"`
	KafkaBroker *KafkaBroker `yaml:"kafka" json:"kafka_broker"`
	Nsq         bool         `yaml:"__nsq" json:"nsq"`
	Rds         bool         `yaml:"__rds" json:"rds"`
	Kfk         bool         `yaml:"__kafka" json:"kafka"`
}

func (b *Broker) Init() {
	if Validated(b.RedisBroker) {
		b.RedisBroker.Init()
		b.Rds = true
	}
	if Validated(b.NsqBroker) {
		b.NsqBroker.Init()
		b.Nsq = true
	}
	if Validated(b.KafkaBroker) {
		b.KafkaBroker.Init()
		b.Kfk = true
	}
}

func (b *Broker) Add(item IItem) bool {
	switch true {
	case Validated(b.RedisBroker):
		b.RedisBroker.Add(item)
	case Validated(b.NsqBroker):
		b.NsqBroker.Add(item)
	case Validated(b.KafkaBroker):
		b.KafkaBroker.Add(item)
	default:
		return false
	}
	return true
}

func (b *Broker) GetBroker() IBroker {
	var broker IBroker
	switch true {
	case Validated(b.RedisBroker):
		broker = b.RedisBroker
	case Validated(b.NsqBroker):
		broker = b.NsqBroker
	default:
		broker = nil
	}
	return broker
}

type Url String

func (u *Url) AddHttp() {
	*u = Url("http://" + u.String())
}

func (u Url) Empty() bool {
	return len(u) == 0
}

func (u *Url) IsHttp() bool {
	return strings.HasPrefix(u.String(), "http://")
}

func (u *Url) IsHttps() bool {
	return strings.HasPrefix(u.String(), "https://")
}

func (u *Url) String() string {
	return string(*u)
}

func (u *Url) Contains(s string) bool {
	if !strings.Contains(s, ",") {
		return strings.Contains(string(*u), s)
	}
	items := strings.Split(s, ",")
	for _, item := range items {
		if strings.Contains(s, item) {
			return true
		}
	}
	return false
}

func (u Url) Host() (string, error) {
	parse, err := url.Parse(string(u))
	if err != nil {
		return "", err
	}
	return parse.Host, nil
}

type Options struct {
	Version  string    `json:"version"`
	AppName  string    `yaml:"kind" json:"app_name"`
	Pages    *Pages    `json:"pages"`
	Broker   *Broker   `json:"broker"`
	Consumer *Consumer `json:"consumer"`
}

type Consumer struct {
	Nsq   *nsqConsumer
	Redis *RedisConsumer
	Kafka *kafkaConsumer
	Limit int
}

func NewNext(arg ...interface{}) (*Next, error) {
	if len(arg) == 0 {
		return &Next{
			G: make(map[string]string),
			R: make(map[string]string),
		}, nil
	}
	for _, a := range arg {
		switch v := a.(type) {
		case map[string]interface{}:
			next := &Next{
				G: make(map[string]string),
				R: make(map[string]string),
				T: make(map[string]string),
				A: make(map[string]*ParserResult),
			}
			err := next.Load(v)
			if err != nil {
				return nil, err
			}
			return next, nil
		default:

		}
	}
	panic("unreach")
}

type Next struct {
	G map[string]string
	R map[string]string
	T map[string]string
	A map[string]*ParserResult
}

func (n *Next) Load(m map[string]interface{}) error {
	if err := mapstructure.Decode(m, n); err != nil {
		return err
	}
	return nil
}

func (n *Next) MergeGr() (result Pattern) {
	result = make(Pattern)
	for k, v := range n.G {
		result[k] = G(v)
	}
	for k, v := range n.R {
		result[k] = R(v)
	}

	for k, v := range n.T {
		result[k] = T(v)
	}
	for k, v := range n.A {
		// v.Key 对应html中指定的元素 如果指定 .next-title元素 v.Value指定对应需要获取的attrib名字
		result[k] = A(_A(v.Key), v.Value.(string))
	}
	return result
}

type Pages struct {
	Labels map[string]*Page
}

type Page struct {
	Next   *Next             `json:"next-parser" yaml:"next-parser"`
	Url    String            `json:"url"`
	Parser G                 `json:"parser"`
	Meta   map[string]string `json:"meta"`
}

func (o *Options) Dumps() (string, error) {
	js, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(js), nil
}

type NsqBroker struct {
	Urls    []string `json:"urls" validate:"required"`
	Topic   string   `json:"Topic" validate:"required"`
	Channel string   `json:"channel" validate:"required"`
	c       *requests.Request
	pushUrl string
	sync.RWMutex
}

func NewNsqConsumer(opt *Consumer) *nsqConsumer {
	return &nsqConsumer{
		Urls:    opt.Nsq.Urls,
		Topic:   opt.Nsq.Topic,
		Channel: opt.Nsq.Channel,
		Limit:   opt.Limit,
	}
}

func NewKafkaConsumer(opt *Consumer) *kafkaConsumer {
	kafka := &kafkaConsumer{
		Addrs: opt.Kafka.Addrs,
		Topic: opt.Kafka.Topic,
		Limit: opt.Limit,
	}
	kafka.Init()
	return kafka
}

func NewRedisConsumer(opt *Consumer) *RedisConsumer {
	r := &RedisConsumer{
		Host:     opt.Redis.Host,
		Password: opt.Redis.Password,
		Db:       opt.Redis.Db,
		Topic:    opt.Redis.Topic,
		Limit:    opt.Limit,
	}
	r.Init()
	return r
}

func randomNsqUrl(v []string) interface{} {
	l := len(v)
	randn := rand.Intn(l)
	if l == 0 {
		return nil
	}
	return v[randn]
}

func (n *NsqBroker) Init() {
	Once(func() {
		n.c = requests.NewRequest()
		log.Infof("nsqd broker inited. host: %s, Topic: %s", n.getPushTopicUrl(), n.Topic)
	})
}

func (n *NsqBroker) getPushTopicUrl() string {
	var params = url.Values{}
	params.Add("Topic", n.Topic)
	n.pushUrl = strings.Join([]string{randomNsqUrl(n.Urls).(string), params.Encode()}, "")
	return n.pushUrl
}

func (n *NsqBroker) Add(item IItem) bool {
	doc, err := item.Dumps()
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		log.Error(err)
		return false
	}
	n.Lock()
	res, err := n.c.PostJson(n.getPushTopicUrl(), doc.String())
	n.Unlock()
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		log.Error(err)
		return false
	}
	log.Debug(res.Text())
	return true
}

type KafkaBroker struct {
	Addrs    []string `json:"addrs"`
	opts     *sarama.Config
	Topic    string `json:"topic"`
	producer sarama.AsyncProducer
	sync.WaitGroup
}

func NewKafkaBroker(addrs []string, topic string) *KafkaBroker {
	k := &KafkaBroker{
		Addrs: addrs,
		Topic: topic,
	}
	k.Init()
	return k
}

func (k *KafkaBroker) Init() {
	var err error
	k.opts = sarama.NewConfig()
	k.opts.Producer.RequiredAcks = sarama.WaitForAll
	k.opts.Producer.Partitioner = sarama.NewRandomPartitioner
	k.producer, err = sarama.NewAsyncProducer(k.Addrs, k.opts)
	if err != nil {
		panic("new kafka producer client error :" + err.Error())
	}
}

func (k *KafkaBroker) Add(item IItem) bool {
	doc, err := item.Dumps()
	if err != nil {
		return false
	}
	msg := &sarama.ProducerMessage{
		Topic: k.Topic,
		Value: sarama.ByteEncoder(doc.String()),
	}
	k.producer.Input() <- msg
	return true
}

type RedisBroker struct {
	Host     string `json:"host" validate:"required"`
	Password string `json:"password"`
	Db       int    `json:"db"`
	Topic    string `json:"Topic" validate:"required"`
	c        *redis.Client
	WaitGroupWrap
}

func (r *RedisBroker) Add(item IItem) bool {
	doc, _ := item.Dumps()
	_, err := r.c.LPush(r.Topic, doc.String()).Result()
	return err == nil
}

func (r *RedisBroker) Init() {
	Once(func() {
		opt := go_utils.NewRedisConf(r.Host, r.Password, r.Db)
		r.c = go_utils.NewRedis(opt)
		log.Info("redis broker inited.")
	})
}

func (o *Options) Item() (String, error) {
	str, err := o.Dumps()
	if err != nil {
		return "", err
	}
	return String(str), nil
}

func NewOptions(path string) (*Options, error) {
	var crawlerOptions Options
	conf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.UnmarshalStrict(conf, &crawlerOptions)
	if err != nil {
		return nil, err
	}
	crawlerOptions.Broker.Init()
	return &crawlerOptions, nil
}

func Host(url string) (string, error) {
	host, err := Url(url).Host()
	if err != nil {
		return "", err
	}
	return host, nil
}

func NewNsqBroker(urls []string, Topic, Channel string) *NsqBroker {
	broker := &NsqBroker{
		Urls:    urls,
		Topic:   Topic,
		Channel: Channel,
		RWMutex: sync.RWMutex{},
	}
	broker.Init()
	return broker
}

func NewRedisBroker(host, password, topic string, db int) *RedisBroker {
	broker := &RedisBroker{
		Host:     host,
		Password: password,
		Db:       db,
		Topic:    topic,
	}
	broker.Init()
	return broker
}
