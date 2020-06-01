package scrapy

import (
	"net/http"
	"sync"
	"time"

	"github.com/Genesis-Palace/requests"
	"gopkg.in/go-playground/validator.v9"
)

type Auth string

type Crawler struct {
	Request *Requests `validate:"required"`
	Cb      func(i IItem)
	Parser  IParser `validate:"required"`
	Item    IItem   `validate:"required"`
	html    String
	Ip      IProxy
	isProxy bool
	ProxyQueue IProxyQueue
	sync.RWMutex
}

func (t *Crawler) Validate() (err error) {
	v := validator.New()
	return v.Struct(t)
}

func (t *Crawler) SetPipelines(cb func(i IItem)) *Crawler {
	t.Lock()
	defer t.Unlock()
	t.Cb = cb
	return t
}

func (t *Crawler) Html() String {
	return t.html
}

func (t *Crawler) ipRecovery(args ...interface{}){
	if !t.isProxy{
		return
	}
	for _, a := range args{
		switch v:= a.(type) {
		case error:
			t.Ip.AddFails()
			log.Error(t.Ip, v)
		default:
			t.Ip.AddSucc()
		}
	}
}

func (t *Crawler) Do() (*Crawler, error) {
	if err := t.Validate(); err != nil {
		t.ipRecovery(err)
		return t, err
	}
	res, err := t.Request.Do()
	if err != nil {
		t.ipRecovery(err)
		return t, err
	}
	t.html = String(res.Text())
	t.Parser.Parser(t.html, t.Item)
	if t.Cb == nil {
		t.Cb = DefaultPipelines
	}
	t.Cb(t.Item)
	t.ipRecovery()
	return t, nil
}

func (t *Crawler) SetHeader(header requests.Header) *Crawler {
	t.Request.SetHeader(header)
	return t
}

func (t *Crawler) SetTimeOut(duration time.Duration) *Crawler {
	t.Request.SetTimeOut(duration)
	return t
}
func (t *Crawler) SetParser(i IParser) *Crawler {
	t.Lock()
	defer t.Unlock()
	t.Parser = i
	return t
}

func (t *Crawler) SetPostJson(json String) *Crawler {
	t.Request.Json(json)
	return t
}

func (t *Crawler) SetMethod(method string) *Crawler {
	t.Request.SetMethod(method)
	return t
}

func (t *Crawler) SetCookies(cookie *http.Cookie) *Crawler {
	t.Request.SetCookies(cookie)
	return t
}

func NewCrawler(url String, args ...interface{}) *Crawler {
	c := &Crawler{
		Request: NewRequest(url),
	}
	for _, arg := range args {
		switch v := arg.(type) {
		case IItem:
			c.Item = v
		}
	}
	c.Item.Add(NewPr("url", url))
	return c
}

func NewProxyCrawler(url String, args ...interface{}) *Crawler {
	c := new(Crawler)
	for _, a := range args {
		switch v := a.(type) {
		case IProxyQueue:
			c.isProxy = true
			clt, ip := v.ProxyClient()
			c.Ip = ip
			c.Request = NewRequest(url, clt)
			c.ProxyQueue = v
		case IItem:
			c.Item = v
		case Auth:
			c.SetHeader(requests.Header{"Proxy-Authorization": string(v)})
		}
	}
	return c
}
