package test

import (
	"fmt"
	crawler "github.com/Genesis-Palace/go-scrapy/scrapy"
	internal "github.com/Genesis-Palace/go-scrapy/scrapy-internal"
	"testing"
	"time"
)

func TestNewMap(t *testing.T) {
	var m = internal.NewMap()
	for i := 0; i < 50; i++ {
		go func(i int) {
			m.Add(map[string]interface{}{string(i): i})
		}(i)
	}
	time.Sleep(time.Second)
	fmt.Println(m.Size())
}

func TestNewList(t *testing.T) {
	var list = internal.NewList()
	for i := 0; i < 50; i++ {
		go func(i int) {
			list.Add(i)
		}(i)
	}
	time.Sleep(time.Second)
	fmt.Println(list.Size())
}

func TestNewMapItem(t *testing.T) {
	var item = internal.NewMap()
	var list = internal.NewList()
	item.Add(map[string]interface{}{"1": 2})
	item.Add(map[string]interface{}{"3": "3"})
	list.Add(1)
	list.Add("3")
	list.Add(5)
	list.Add(7)
	item.Add(map[string]interface{}{"5": list.Items()})
	fmt.Println(item.Dumps())
}

func TestLoadOptions(t *testing.T) {
	path := "../apps/config/test_list.yaml"
	options, err := crawler.NewOptions(path)
	if err != nil {
		t.Error(err)
	}
	t.Log(options)
}

func TestStringItem(t *testing.T) {
	var str internal.String
	empty := str.Empty()
	if !empty {
		t.Error("str len errors. str is not empty")
	}
	t.Log("str is empty, len = 0")
	str = "tag--newids"
	has := str.HasPrefix("tag--")
	if !has {
		t.Error("tag-- is not in str")
	}
	fmt.Println(str.String(), str.Hash())
}

func TestNewTasks(t *testing.T) {
	var urls = []internal.String{
		"https://www.toutiao.com/a6817599019259265539/",
		"https://www.toutiao.com/a6819580497153229320/",
		"https://www.toutiao.com/a6817772303032517128/",
		//"http://www.toutiaonews.com/",
	}
	for _, url := range urls {
		go func(url internal.String) {
			var item = internal.NewMap()
			item.Add(internal.NewPr("url", url))
			item.Add(internal.NewPr("ts", time.Now().Unix()))
			var parser = internal.NewMixdParser(internal.Pattern{
				"title":     internal.G("head title"),
				"tag":       internal.R(`chineseTag: '(.*?)'`),
				"group_id":  internal.R(`groupId: '(.*?)'`),
				"publisher": internal.R(`name: '(.*?)'`),
				"uid":       internal.R(`uid: '(.*?)'`),
			})
			task := crawler.NewCrawler(url, item)
			task.SetTimeOut(3).SetParser(parser).Do()
		}(url)
	}
	time.Sleep(time.Second)
}

func TestNewCrawler(t *testing.T) {
	var uc = make(chan internal.String, 100)
	go func() {
		var url = internal.String("http://www.toutiaonews.com")
		var parser = internal.NewGoQueryParser(".newsList li dl dt a")
		var item = internal.NewMap()
		c := crawler.NewCrawler(url, item)
		c.SetTimeOut(3 * time.Second).SetParser(parser).Do()
		host, _ := crawler.Url(url).Host()
		fmt.Println(len(item.Items()))
		for _, v := range item.Items()["href"].([]interface{}) {
			if _, err := crawler.Host(v.(string)); err != nil {
				continue
			}
			uc <- internal.String(fmt.Sprintf("http://%s/%s", host, v))
		}
		close(uc)
	}()
	for u := range uc {
		var parser = internal.NewMixdParser(internal.Pattern{
			"title":     internal.G(".article-body h1"),
			"date":      internal.R(`时间：([0-9].*[0-9])`),
			"recommend": internal.G(".related-recul li a"),
		})
		var item = internal.NewMap()
		c := crawler.NewCrawler(u, item)
		c.SetTimeOut(3 * time.Second).SetParser(parser).Do()
		break
	}
}
