package test

import (
	crawler "github.com/Genesis-Palace/go-scrapy/scrapy"
	"path/filepath"
	"testing"
	"time"
)

func TestNewMap(t *testing.T) {
	var m = crawler.NewMap()
	for i := 0; i < 50; i++ {
		go func(i int) {
			m.Add(map[string]interface{}{string(i): i})
		}(i)
	}
	time.Sleep(time.Second)
	t.Log(m.Size())
}

func TestNewList(t *testing.T) {
	var list = crawler.NewList()
	for i := 0; i < 50; i++ {
		go func(i int) {
			list.Add(i)
		}(i)
	}
	time.Sleep(time.Second)
	t.Log(list.Size())
}

func TestNewMapItem(t *testing.T) {
	var item = crawler.NewMap()
	var list = crawler.NewList()
	item.Add(map[string]interface{}{"1": 2})
	item.Add(map[string]interface{}{"3": "3"})
	list.Add(1)
	list.Add("3")
	list.Add(5)
	list.Add(7)
	item.Add(map[string]interface{}{"5": list.Items()})
	t.Log(item.Dumps())
}

func TestLoadOptions(t *testing.T) {
	path := "../producer.yaml"
	t.Log(filepath.Abs("."))
	options, err := crawler.NewOptions(path)
	if err != nil {
		t.Error(err)
	}
	t.Log(options)
}

func TestStringItem(t *testing.T) {
	var str crawler.String
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
	t.Log(str.String(), str.Hash())
}

func TestNewTasks(t *testing.T) {
	var urls = []crawler.String{
		"https://www.toutiao.com/a6817599019259265539/",
		"https://www.toutiao.com/a6819580497153229320/",
		"https://www.toutiao.com/a6817772303032517128/",
		//"http://www.toutiaonews.com/",
	}
	for _, url := range urls {
		go func(url crawler.String) {
			var item = crawler.NewMap()
			item.Add(crawler.NewPr("url", url))
			item.Add(crawler.NewPr("ts", time.Now().Unix()))
			var parser = crawler.NewMixdParser(crawler.Pattern{
				"title":     crawler.G("head title"),
				"tag":       crawler.R(`chineseTag: '(.*?)'`),
				"group_id":  crawler.R(`groupId: '(.*?)'`),
				"publisher": crawler.R(`name: '(.*?)'`),
				"uid":       crawler.R(`uid: '(.*?)'`),
			})
			task := crawler.NewCrawler(url, item)
			task.SetTimeOut(3).SetParser(parser).Do()
		}(url)
	}
	time.Sleep(time.Second)
}

//func TestNewCrawler(t *testing.T) {
//	var uc = make(chan crawler.String, 100)
//	go func() {
//		var url = crawler.String("http://www.toutiaonews.com")
//		var parser = crawler.NewGoQueryParser(".newsList li dl dt a")
//		var item = crawler.NewMap()
//		c := crawler.NewCrawler(url, item)
//		c.SetTimeOut(3 * time.Second).SetParser(parser).Do()
//		host, _ := crawler.Url(url).Host()
//		for _, v := range item.Items()["href"].([]interface{}) {
//			if _, err := crawler.Host(v.(string)); err != nil {
//				continue
//			}
//			uc <- crawler.String(fmt.Sprintf("http://%s/%s", host, v))
//		}
//		close(uc)
//	}()
//	for u := range uc {
//		var parser = crawler.NewMixdParser(crawler.Pattern{
//			"title":     crawler.G(".article-body h1"),
//			"date":      crawler.R(`时间：([0-9].*[0-9])`),
//			"recommend": crawler.G(".related-recul li a"),
//		})
//		var item = crawler.NewMap()
//		c := crawler.NewCrawler(u, item)
//		c.SetTimeOut(3 * time.Second).SetParser(parser).Do()
//		t.Log(item.Items())
//	}
//}
