package scrapy

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2"
	"math/rand"
	"time"

	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/go-bongo/bongo"
	"github.com/go-redis/redis"
	"gopkg.in/mgo.v2/bson"
)

type MongoClient struct {
	config     []*bongo.Config
	connection []*bongo.Connection
	colCh      chan string
}

func (m *MongoClient) instance() *bongo.Connection {
	randn := rand.Intn(len(m.connection))
	if len(m.connection) == 0 {
		return nil
	}
	if len(m.colCh) == 0 {
		log.Warning("Collection name is not specified")
		return nil
	}
	return m.connection[randn]
}

func (m *MongoClient) Add(doc bongo.Document) bool {
	c := m.instance()
	if c == nil {
		return false
	}
	return nil == c.Collection(<-m.colCh).Save(doc)
}

func (m *MongoClient) Count(m2 bson.M) (int, error) {
	c := m.instance()
	if c == nil {
		return 0, nil
	}
	return c.Session.DB(c.Config.Database).C(<-m.colCh).Count()
}

func (m *MongoClient) Collection(col string) *MongoClient {
	m.colCh <- col
	return m
}

func (m *MongoClient) FindOne(query interface{}, result interface{}) error {
	c := m.instance()
	if c == nil {
		return errors.New("collection is nil")
	}
	err := c.Collection(<-m.colCh).FindOne(query, &result)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (m *MongoClient) Remove(query bson.M) bool {
	c := m.instance()
	if c == nil {
		return false
	}
	changeInfo, err := c.Collection(<-m.colCh).Delete(query)
	if err == nil {
		log.Debugf("remove docs, matched: %d, removed: %d", changeInfo.Matched, changeInfo.Removed)
		return true
	}
	return false
}

// 删除单条doc, 如果doc 包含BeforeDelete 和 AfterDeleteHook 则触发.
func (m *MongoClient) Del(doc bongo.Document) {
	c := m.instance()
	if c == nil {
		return
	}
	err := m.instance().Collection(<-m.colCh).DeleteDocument(doc)
	if err != nil {
		log.Error(err)
	}
}

func (m *MongoClient) Find(query interface{}) *bongo.ResultSet {
	c := m.instance()
	if c == nil {
		return nil
	}
	return c.Collection(<-m.colCh).Find(query)
}

func (m *MongoClient) Pipe(args ...bson.M) *mgo.Pipe {
	c := m.instance()
	if c == nil {
		return nil
	}
	return c.Session.DB(c.Config.Database).C(<-m.colCh).Pipe(args)
}

func (m *MongoClient) Init() {
	for _, conf := range m.config {
		connection, err := bongo.Connect(conf)
		if err != nil {
			log.Error(err)
			continue
		}
		m.connection = append(m.connection, connection)
	}
}

func NewMongoClient(config []*bongo.Config) *MongoClient {
	c := &MongoClient{
		config: config,
		colCh:  make(chan string, 1),
	}
	c.Init()
	return c
}

type RedisClient struct {
	redisClients []*redis.Client
	opts         []*redis.Options
}

func (r *RedisClient) Instance() *redis.Client {
	var seed = rand.Intn(len(r.redisClients))
	return r.redisClients[seed]
}

func (r *RedisClient) Lpop(key string) (string, error) {
	return r.Instance().SPop(key).Result()
}

func (r *RedisClient) Pipelines(fn func(pipeliner redis.Pipeliner) error) error {
	cmd, err := r.Instance().Pipelined(fn)
	log.Info(cmd)
	return err
}

func (r *RedisClient) SCard(key string) (int64, error) {
	return r.Instance().SCard(key).Result()
}

func (r *RedisClient) LLen(key string) (int64, error) {
	return r.Instance().LLen(key).Result()
}

func (r *RedisClient) Publish(channels string, msg interface{}) error {
	_, err := r.Instance().Pipelined(func(pipeliner redis.Pipeliner) error {
		return pipeliner.Publish(channels, msg).Err()
	})
	return err
}

func (r *RedisClient) Sub(channels ...string) <-chan *redis.Message {
	sub := r.Instance().Subscribe(channels...)
	return sub.Channel()
}

func (r *RedisClient) SPopN(key string, count int64) ([]string, error) {
	return r.Instance().SPopN(key, count).Result()
}

func (r *RedisClient) Sorted(key string, sort *redis.Sort) *redis.StringSliceCmd {
	return r.Instance().Sort(key, sort)
}

func (r *RedisClient) Expire(key string, duration time.Duration) {
	r.Instance().Expire(key, duration)
}

func (r *RedisClient) Incr(key string) int64 {
	return r.Instance().Incr(key).Val()
}

func (r *RedisClient) Existed(key string) bool {
	_, err := r.Instance().Get(key).Int()
	return err != nil
}

func (r *RedisClient) SIsMember(key, id string) bool {
	return r.Instance().SIsMember(key, id).Val()
}

func (r *RedisClient) Lpush(key string, val interface{}) {
	r.Instance().LPush(key, val)
}

func (r *RedisClient) MaxKeyCount(key string, max int) bool {
	val, err := r.Instance().Get(key).Int()
	if err != nil {
		if err == redis.Nil {
			return true
		}
	}
	if val <= max {
		return true
	}
	return false
}

func NewRedis(args ...*redis.Options) (*RedisClient, error) {
	var redisClients []*redis.Client
	var opts []*redis.Options
	if len(args) == 0 {
		return nil, fmt.Errorf("new redis args is empty. please enter redis options")
	}
	for _, arg := range args {
		redisClients = append(redisClients, go_utils.NewRedis(arg))
		opts = append(opts, arg)
	}
	return &RedisClient{
		redisClients: redisClients,
		opts:         opts,
	}, nil
}

func NewRedisCluster(ips []string, password string) *redis.ClusterClient {
	cluster := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    ips,
		Password: password,
	})
	return cluster
}
