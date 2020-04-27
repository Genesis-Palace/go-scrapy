package test

import (
	"github.com/Genesis-Palace/requests"
	"go-scrapy/internal"
	"testing"
	"time"
)

func TestNewRequest(t *testing.T) {
	for i := 0; i <= 5; i++ {
		go func() {
			var url internal.String = "http://httpbin.org/headers"
			req := internal.NewRequest(url)
			req.SetHeader(requests.Header{
				"Host": "www.abuyun.com",
				"Referer": "https://www.abuyun.com/http-proxy/dyn-manual.html",
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel …) Gecko/20100101 Firefox/75.0",
			},)
			resp, err := req.Do()
			if err != nil {
				t.Error(err)
			}
			t.Log(resp.Text())
		}()
	}
	time.Sleep(time.Second)
}

func TestNewProxyRequest(t *testing.T) {
	var url internal.String = "http://httpbin.org/ip"
	req := internal.NewRequest(url, internal.Use)
	resp, err := req.Do()
	if err != nil {
		t.Error(err)
	}
	t.Log(resp.Text())
}

func TestNewRequestPost(t *testing.T) {
	var url internal.String = "http://httpbin.org/post"
	var body = internal.NewMap()
	var item = make(map[string]interface{})
	item["a"] = 1
	item["b"] = 2
	item["c"] = 3
	body.Add(item)
	js, err := body.Dumps()
	if err != nil {
		t.Error(err)
	}
	req := internal.NewRequest(url)
	resp, err := req.SetMethod(internal.POSTJSON).Json(js).SetTimeOut(5).Do()
	if err != nil {
		t.Error(err)
	}
	t.Log(resp.Text())
}

func TestNewRequestArgs(t *testing.T) {
	var url internal.String = "http://httpbin.org/ip"
	req := internal.NewRequest(
		url,
		requests.Header{
			"Host": "www.abuyun.com",
			"Connection": "keep-alive",
			"Referer": "https://www.abuyun.com/http-proxy/dyn-manual.html",
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel …) Gecko/20100101 Firefox/75.0",
		},
		internal.Use,
		time.Duration(3) * time.Second,
	)
	resp, err := req.Do()
	if err != nil{
		t.Error(err)
	}
	t.Log(resp.Text())
	for k, v := range resp.R.Request.Header{
		t.Log(k, v)
	}
}
