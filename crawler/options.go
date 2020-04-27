package crawler

import (
	"encoding/json"
	"github.com/Genesis-Palace/go-utils"
	"github.com/Genesis-Palace/requests"
	"github.com/go-redis/redis"
	"github.com/goinggo/mapstructure"
	"go-srapy/internal"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	log = go_utils.Log()
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

type BrokerInterfaceI interface {
	Init()
	Add(item internal.ItemInterfaceI) bool
}

type Broker struct {
	NsqBroker   *NsqBroker   `yaml:"nsq" json:"nsq_broker"`
	RedisBroker *RedisBroker `yaml:"redis" json:"redis_broker"`
	Nsq         bool         `yaml:"__nsq" json:"nsq"`
	Rds         bool         `yaml:"__rds" json:"rds"`
}

func (b *Broker) Init() {
	if internal.Validated(b.RedisBroker) {
		b.RedisBroker.Init()
		b.Rds = true
	}
	if internal.Validated(b.NsqBroker) {
		b.NsqBroker.Init()
		b.Nsq = true
	}
}

func (b *Broker) Add(item internal.ItemInterfaceI) bool {
	switch true {
	case internal.Validated(b.RedisBroker):
		b.RedisBroker.Add(item)
	case internal.Validated(b.NsqBroker):
		b.NsqBroker.Add(item)
	default:
		return false
	}
	return true
}

func (b *Broker) GetBroker() BrokerInterfaceI {
	var broker BrokerInterfaceI
	switch true {
	case internal.Validated(b.RedisBroker):
		broker = b.RedisBroker
	case internal.Validated(b.NsqBroker):
		broker = b.NsqBroker
	default:
		broker = nil
	}
	return broker
}

type Url internal.String

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
	Limit int
}

type Next struct {
	G map[string]string
	R map[string]string
}

func (n *Next) Load(m map[string]interface{}) error{
	if err := mapstructure.Decode(m, n); err != nil {
		return err
	}
	return nil
}

func (n *Next) MergeGr() (result internal.Pattern){
	result = make(internal.Pattern)
	for k, v := range n.G {
		result[k] = internal.G(v)
	}
	for k, v := range n.R {
		result[k] = internal.R(v)
	}
	return result
}

type Pages struct {
	Labels map[string]*Page
}

type Page struct {
	Next   *Next             `json:"next-parser" yaml:"next-parser"`
	Url    internal.String   `json:"url"`
	Parser internal.G        `json:"parser"`
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
	Topic   string   `json:"topic" validate:"required"`
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
	internal.Once(func() {
		n.c = requests.NewRequest()
		log.Infof("nsqd broker inited. host: %s, topic: %s", n.getPushTopicUrl(), n.Topic)
	})
}

func (n *NsqBroker) getPushTopicUrl() string {
	var params = url.Values{}
	params.Add("topic", n.Topic)
	n.pushUrl = strings.Join([]string{randomNsqUrl(n.Urls).(string), params.Encode()}, "")
	return n.pushUrl
}

func (n *NsqBroker) Add(item internal.ItemInterfaceI) bool {
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

type RedisBroker struct {
	Host     string `json:"host" validate:"required"`
	Password string `json:"password"`
	Db       int    `json:"db"`
	Topic    string `json:"topic" validate:"required"`
	c        *redis.Client
	internal.WaitGroupWrap
}

func (r *RedisBroker) Add(item internal.ItemInterfaceI) bool {
	doc, _ := item.Dumps()
	_, err := r.c.LPush(r.Topic, doc.String()).Result()
	if err != nil {
		return false
	}
	return true
}

func (r *RedisBroker) Init() {
	internal.Once(func() {
		opt := go_utils.NewRedisConf(r.Host, r.Password, r.Db)
		r.c = go_utils.NewRedis(opt)
		log.Info("redis broker inited.")
	})
}

func (o *Options) Item() (internal.String, error) {
	str, err := o.Dumps()
	if err != nil {
		return "", err
	}
	return internal.String(str), nil
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

