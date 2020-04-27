package crawler

import (
	"github.com/Genesis-Palace/go-scrapy/internal"
	"github.com/Genesis-Palace/requests"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"sync"
	"time"
)

type Crawler struct {
	Request *internal.Requests `validate:"required"`
	Cb      func(i internal.ItemInterfaceI)
	Parser  internal.ParserInterfaceI `validate:"required"`
	Item    internal.ItemInterfaceI   `validate:"required"`
	sync.RWMutex
}

func (t *Crawler) Validate() (err error) {
	v := validator.New()
	return v.Struct(t)
}

func (t *Crawler) SetPipelines(cb func(i internal.ItemInterfaceI)) {
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
	t.Parser.Parser(internal.String(internal.AutoGetHtmlEncode(res.Text())), t.Item)
	if t.Cb == nil {
		t.Cb = internal.DefaultPipelines
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
func (t *Crawler) SetParser(i internal.ParserInterfaceI) *Crawler {
	t.Lock()
	defer t.Unlock()
	t.Parser = i
	return t
}

func (t *Crawler) SetPostJson(json internal.String) *Crawler {
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

func NewCrawler(url internal.String, item internal.ItemInterfaceI) *Crawler {
	item.Add(internal.NewPr("url", url))
	return &Crawler{
		Request: internal.NewRequest(url),
		Item:    item,
	}
}

func NewProxyCrawler(url internal.String, proxy *internal.AbuyunProxy, item internal.ItemInterfaceI) *Crawler {
	return &Crawler{
		Request: internal.NewRequest(url, internal.Use, proxy),
		Item:    item,
	}
}
