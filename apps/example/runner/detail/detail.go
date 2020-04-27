package detail

import (
	"errors"
	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/nsqio/go-nsq"
	"go-srapy/crawler"
	"go-srapy/internal"
)

var (
	log = go_utils.Log()
)

type IHandler struct{}

func (i *IHandler) HandleMessage(msg *nsq.Message) error {
	var m = internal.NewMap()
	err := m.Load(msg.Body)
	if err != nil{
		return err
	}
	urls := m.Pop("href")
	host, _ := crawler.Url(m.Get("url").(string)).Host()
	if urls == nil{
		return errors.New("list urls is nil. please check seeds urls.")
	}
	var next crawler.Next
	nextInterface := m.Get("next")
	if nextInterface == nil{
		return errors.New("next parser struct is not define.")
	}
	err = next.Load(nextInterface.(map[string]interface{}))
	if err != nil{
		return err
	}
	nextInfo := next.MergeGr()
	log.Info(nextInfo)
	for _, item := range urls.([]interface{}){
		var u = crawler.Url(item.(string))
		var result = internal.NewMap()
		var parser = internal.NewMixdParser(nextInfo)
		url := crawler.Url(host + u.String())
		if !url.IsHttp(){
			url.AddHttp()
		}
		crawler.NewCrawler(internal.String(url), result).SetParser(parser).SetTimeOut(3).Do()
	}
	return nil
}

func NsqQueueMain(options *crawler.Consumer) {
	crawler.NewNsqConsumer(options).SetHandler(new(IHandler)).Run()
}

func RedisQueueMain(options *crawler.Consumer) {
	crawler.NewRedisConsumer(options).SetHandler(new(IHandler)).Run()
}
