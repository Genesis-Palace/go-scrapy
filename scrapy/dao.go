package scrapy

import (
	"github.com/go-bongo/bongo"
	"math/rand"
	"sync"
)

type MongoClient struct {
	config     []*bongo.Config
	connection []*bongo.Connection
	sync.RWMutex
}

func (m *MongoClient) Instance() *bongo.Connection {
	randn := rand.Intn(len(m.connection))
	if len(m.connection) == 0 {
		return nil
	}
	return m.connection[randn]
}

func (m *MongoClient) Add(item *ParserResult) bool {
	return nil == m.Instance().Collection(item.Key).Save(item.Value.(bongo.Document))
}

func (m *MongoClient) Init() {
	m.Lock()
	defer m.Unlock()
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
	}
	c.Init()
	return c
}
