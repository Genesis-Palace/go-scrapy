package scrapy

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Genesis-Palace/requests"
	"github.com/ymzuiku/hit"
)

type BrowserName String

const (
	GET                          = "get"
	POST                         = "post"
	POSTJSON                     = "post-json"
	DefaultTimeOut time.Duration = 1
)

func NewAbutunProxy(appid, secret, proxyServer string) *AbuyunProxy {
	return &AbuyunProxy{
		AppID:       appid,
		AppSecret:   secret,
		ProxyServer: proxyServer,
	}
}

type AbuyunProxy struct {
	AppID       string
	AppSecret   string
	ProxyServer string
}

func (p AbuyunProxy) pullIps() {
	panic("implement me")
}

func (p AbuyunProxy) Init() {
	panic("implement me")
}

func (p AbuyunProxy) Get() IProxy {
	panic("implement me")
}

func (p AbuyunProxy) Put(proxy IProxy) {
	panic("implement me")
}

func (p AbuyunProxy) ProxyClient() (*http.Client, IProxy) {
	proxyUrl, _ := url.Parse("http://" + p.AppID + ":" + p.AppSecret + "@" + p.ProxyServer)
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}, nil
}

type IClient interface {
	Get(url String, args ...interface{}) (*requests.Response, error)
	PostJson(url, js String, args ...interface{}) (*requests.Response, error)
	SetTimeOut(duration time.Duration)
	SetHeaders(header requests.Header)
}

type DefaultClient struct {
	c *requests.Request
	sync.RWMutex
}

func (d *DefaultClient) SetTimeOut(duration time.Duration) {
	d.Lock()
	defer d.Unlock()
	d.c.SetTimeout(duration)
}
func (d *DefaultClient) Get(url String, args ...interface{}) (resp *requests.Response, err error) {
	d.Lock()
	defer d.Unlock()
	return d.c.Get(url.String(), args)
}

func (d *DefaultClient) PostJson(url, js String, args ...interface{}) (*requests.Response, error) {
	d.Lock()
	defer d.Unlock()
	return d.c.PostJson(url.String(), js.String(), args)
}

func (d *DefaultClient) SetHeaders(header requests.Header) {
	for k, v := range header {
		d.c.Header.Add(k, v)
	}
}

type ProxyClient struct {
	c *requests.Request
	sync.RWMutex
}

func (p *ProxyClient) SetHeaders(header requests.Header) {
	for k, v := range header {
		p.c.Header.Add(k, v)
	}
}

func (p *ProxyClient) SetTimeOut(duration time.Duration) {
	p.Lock()
	defer p.Unlock()
	p.c.SetTimeout(duration)
}

func (p *ProxyClient) Get(url String, args ...interface{}) (resp *requests.Response, err error) {
	p.Lock()
	defer p.Unlock()
	return p.c.Get(url.String(), args)
}

func (p *ProxyClient) PostJson(url, js String, args ...interface{}) (*requests.Response, error) {
	p.Lock()
	defer p.Unlock()
	return p.c.PostJson(url.String(), js.String(), args)
}

type Requests struct {
	Url     String
	headers requests.Header
	cookies *http.Cookie
	method  String
	timeout time.Duration
	json    String
	c       IClient
	sync.RWMutex
}

func (r *Requests) Json(js String) *Requests {
	r.Lock()
	r.json = js
	r.Unlock()
	return r
}

func (r *Requests) SetMethod(method string) *Requests {
	r.Lock()
	r.method = String(method)
	r.Unlock()
	return r
}

func (r *Requests) SetTimeOut(timeout time.Duration) *Requests {
	r.Lock()
	r.timeout = timeout
	r.c.SetTimeOut(timeout)
	r.Unlock()
	return r
}

func (r *Requests) SetHeader(headers requests.Header) *Requests {
	r.Lock()
	defer r.Unlock()
	r.c.SetHeaders(headers)
	return r
}

func (r *Requests) SetCookies(cookie *http.Cookie) *Requests {
	r.Lock()
	r.cookies = cookie
	r.Unlock()
	return r
}

func (r *Requests) Do() (resp *requests.Response, err error) {
	if r.method.Empty() {
		r.method = GET
	}
	durtion := hit.If(r.timeout > DefaultTimeOut, r.timeout, DefaultTimeOut).(time.Duration)
	r.c.SetTimeOut(durtion)
	switch r.method {
	case GET:
		return r.c.Get(r.Url)
	case POSTJSON:
		return r.c.PostJson(r.Url, r.json)
	case POST:
	}
	panic("unreach")
}

type Response struct {
	*requests.Response
}

func NewRequest(url String, args ...interface{}) *Requests {
	var req = &Requests{
		Url:     url,
		c:       NewDefaultClient(),
	}
	for _, arg := range args {
		switch v := arg.(type) {
		case requests.Header:
			req.SetHeader(v)
		case *http.Cookie:
			req.SetCookies(v)
		case time.Duration:
			req.SetTimeOut(v)
		case *http.Client:
			req.c = newProxyClient(v)
		case IClient:
			req.c = v
		default:
		}
	}
	return req
}

func NewDefaultClient() IClient {
	return &DefaultClient{
		c:       requests.Requests(),
	}
}

func newProxyClient(c *http.Client) IClient {
	client := requests.Requests()
	client.Client = c
	return &ProxyClient{
		c:       client,
	}
}
