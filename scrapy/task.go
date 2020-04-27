package scrapy

import (
	"github.com/Genesis-Palace/go-scrapy/scrapy-internal"
	"github.com/Genesis-Palace/requests"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"sync"
	"time"
)

type Crawler struct {
	Request *scrapy_internal.Requests `validate:"required"`
	Cb      func(i scrapy_internal.ItemInterfaceI)
	Parser  scrapy_internal.ParserInterfaceI `validate:"required"`
	Item    scrapy_internal.ItemInterfaceI   `validate:"required"`
	sync.RWMutex
}

func (t *Crawler) Validate() (err error) {
	v := validator.New()
	return v.Struct(t)
}

func (t *Crawler) SetPipelines(cb func(i scrapy_internal.ItemInterfaceI)) {
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
	t.Parser.Parser(scrapy_internal.String(scrapy_internal.AutoGetHtmlEncode(res.Text())), t.Item)
	if t.Cb == nil {
		t.Cb = scrapy_internal.DefaultPipelines
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
func (t *Crawler) SetParser(i scrapy_internal.ParserInterfaceI) *Crawler {
	t.Lock()
	defer t.Unlock()
	t.Parser = i
	return t
}

func (t *Crawler) SetPostJson(json scrapy_internal.String) *Crawler {
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

func NewCrawler(url scrapy_internal.String, item scrapy_internal.ItemInterfaceI) *Crawler {
	item.Add(scrapy_internal.NewPr("url", url))
	return &Crawler{
		Request: scrapy_internal.NewRequest(url),
		Item:    item,
	}
}

func NewProxyCrawler(url scrapy_internal.String, proxy *scrapy_internal.AbuyunProxy, item scrapy_internal.ItemInterfaceI) *Crawler {
	return &Crawler{
		Request: scrapy_internal.NewRequest(url, scrapy_internal.Use, proxy),
		Item:    item,
	}
}
