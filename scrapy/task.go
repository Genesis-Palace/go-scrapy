package scrapy

import (
	"github.com/Genesis-Palace/requests"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"sync"
	"time"
)

type Crawler struct {
	Request *Requests `validate:"required"`
	Cb      func(i ItemInterfaceI)
	Parser  ParserInterfaceI `validate:"required"`
	Item    ItemInterfaceI   `validate:"required"`
	html    String
	sync.RWMutex
}

func (t *Crawler) Validate() (err error) {
	v := validator.New()
	return v.Struct(t)
}

func (t *Crawler) SetPipelines(cb func(i ItemInterfaceI)) *Crawler{
	t.Lock()
	defer t.Unlock()
	t.Cb = cb
	return t
}

func (t *Crawler) Html() String {
	return t.html
}

func (t *Crawler) Do() *Crawler {
	if err := t.Validate(); err != nil {
		log.Error(err)
		return t
	}
	res, err := t.Request.Do()
	if err != nil {
		return t
	}
	t.html = String(res.Text())
	t.Parser.Parser(t.html, t.Item)
	if t.Cb == nil {
		t.Cb = DefaultPipelines
	}
	t.Cb(t.Item)
	return t
}

func (t *Crawler) SetHeader(header requests.Header) *Crawler {
	t.Request.SetHeader(header)
	return t
}

func (t *Crawler) SetTimeOut(duration time.Duration) *Crawler {
	t.Request.SetTimeOut(duration)
	return t
}
func (t *Crawler) SetParser(i ParserInterfaceI) *Crawler {
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
	for _, arg := range args{
		switch arg.(type) {
		case ItemInterfaceI:
			c.Item = arg.(ItemInterfaceI)
		}
	}
	c.Item.Add(NewPr("url", url))
	return c
}

func NewProxyCrawler(url String, proxy *AbuyunProxy, item ItemInterfaceI) *Crawler {
	return &Crawler{
		Request: NewRequest(url, proxy),
		Item:    item,
	}
}
