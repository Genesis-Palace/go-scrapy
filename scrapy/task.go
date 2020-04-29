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
	sync.RWMutex
}

func (t *Crawler) Validate() (err error) {
	v := validator.New()
	return v.Struct(t)
}

func (t *Crawler) SetPipelines(cb func(i ItemInterfaceI)) {
	t.Lock()
	defer t.Unlock()
	t.Cb = cb
}

func (t *Crawler) Do() {
	if err := t.Validate(); err != nil {
		log.Error(err)
		return
	}
	res, err := t.Request.Do()
	if err != nil {
		log.Error(err)
		return
	}
	t.Parser.Parser(String(AutoGetHtmlEncode(res.Text())), t.Item)
	if t.Cb == nil {
		t.Cb = DefaultPipelines
	}
	t.Cb(t.Item)
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

func NewCrawler(url String, item ItemInterfaceI) *Crawler {
	item.Add(NewPr("url", url))
	return &Crawler{
		Request: NewRequest(url),
		Item:    item,
	}
}

func NewProxyCrawler(url String, proxy *AbuyunProxy, item ItemInterfaceI) *Crawler {
	return &Crawler{
		Request: NewRequest(url, Use, proxy),
		Item:    item,
	}
}
