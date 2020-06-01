package scrapy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	ProxIpMaxExpiredTime float64 = 5
)

type IProxyQueue interface {
	pullIps()
	Init()
	Get() IProxy
	Put(IProxy)
	ProxyClient() (*http.Client, IProxy)
}

type ProxyQueue struct {
	Queue   chan IProxy
	PullUrl string
	MaxCaps int
	Sleep   time.Duration
	sync.RWMutex
}

func (p *ProxyQueue) ProxyClient() (*http.Client, IProxy) {
	ip := p.Get()
	urlproxy, _ := url.Parse(ip.ProxyUrl())
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(urlproxy),
		},
	}, ip
}

func (q *ProxyQueue) Get() IProxy {
	for {
		ip := <-q.Queue
		if ip.isAvailable() {
			q.Queue <- ip
			return ip
		}
	}
}

func (q *ProxyQueue) Put(ip IProxy) {
	q.Queue <- ip
}

func (q *ProxyQueue) pullIps() {
	for {
		if len(q.Queue) > q.MaxCaps/2 {
			time.Sleep(time.Second)
			continue
		}
		var item = NewMap()
		_, _ = NewCrawler(String(q.PullUrl), item).SetParser(NewJsonParser()).Do()
		data := item.Pop("data")
		var proxyList []map[string]interface{}
		if data == nil {
			continue
		}
		js, err := json.Marshal(data.(map[string]interface{})["proxy_list"])
		if err != nil {
			continue
		}
		err = json.Unmarshal(js, &proxyList)
		if err != nil {
			log.Error(err)
			time.Sleep(q.Sleep)
			continue
		}
		for _, item := range proxyList {
			q.Queue <- NewProxyIp(item["ip"].(string), int(item["port"].(float64)))
		}
		time.Sleep(q.Sleep)
	}
}

func (q *ProxyQueue) Init() {
	go q.pullIps()
}

type IProxy interface {
	AddFails()
	AddSucc()
	Expired() bool
	isAvailable() bool
	ProxyUrl() string
}

func NewProxyIp(host string, port int) *ProxyIp {
	ip := &ProxyIp{
		Host:       host,
		Port:       port,
		CreateTime: time.Now(),
		Failures:   0,
		Used:       0,
		Available:  true,
	}
	return ip
}

type ProxyIp struct {
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	CreateTime time.Time `json:"create_time"`
	Failures   int       `json:"failures"`
	Used       int       `json:"used"`
	Available  bool      `json:"available"`
	sync.RWMutex
}

func (p *ProxyIp) ProxyUrl() string {
	return fmt.Sprintf("http://%s:%d", p.Host, p.Port)
}

func (p *ProxyIp) Expired() bool {
	return time.Since(p.CreateTime).Minutes() > ProxIpMaxExpiredTime
}

func (p *ProxyIp) AddSucc() {
	p.Lock()
	defer p.Unlock()
	p.Used += 1
}

func (p *ProxyIp) AddFails() {
	p.Lock()
	defer p.Unlock()
	p.Failures += 1
}

func (p *ProxyIp) isAvailable() bool {
	if p.Expired() || p.Failures >= 1 {
		p.Available = false
	}
	return p.Available
}
