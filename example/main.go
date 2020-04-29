package main

import (
	"errors"
	"fmt"
	"github.com/Genesis-Palace/go-scrapy/scrapy"
	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/nsqio/go-nsq"
)

var (
	log = go_utils.Log()
	broker = scrapy.NewRedisBroker("127.0.0.1:6379", "", "test-redis-broker", 0)
)

// 代理设置方法
func example3(){
	// 目前代理只支持阿布云. 没有提供client的相关接口
	var item = scrapy.NewMap()
	var proxy = scrapy.NewAbutunProxy("appid", "secret", "proxyserver")
	var parser = scrapy.NewGoQueryParser("head title")
	var url scrapy.String = "https://www.toutiao.com/i6790992050591367684"
	scrapy.NewProxyCrawler(url, proxy, item).SetParser(parser).SetTimeOut(1).Do()
	fmt.Println(item.Items())
}

func BrokerDemo(){
	var url scrapy.String = "https://www.toutiao.com/i6820706041991266827"
	var item = scrapy.NewMap()
	var parser = scrapy.NewMixdParser(scrapy.Pattern{
		"title": scrapy.G("head title"),
		"abstract": scrapy.R("abstract: '(.*?)'"),
	})
	scrapy.NewCrawler(url, item).SetTimeOut(1).SetParser(parser).Do()
	if item.Empty(){
		log.Error(errors.New("item is empty."))
		return
	}
	log.Info(item)
	broker.Add(item)
	/*
	127.0.0.1:6379> lpop test-redis-topic
		{"abstract":"\"继我国15式轻型坦克之后，美军也不甘落后，在最近推出了一款名为MPF的新式轻型坦克，那么这款轻型坦克性能到底如何？\"","title":"美军新轻坦闪亮登场，中国15式坦克终于迎来对手？对比过后直摇头","url":"https://www.toutiao.com/i6820706041991266827"}
	 */
}


// 当任务需要批量配置化时推荐使用该方案.
// 通过配置文件配置好首次采集的种子页面, 通过种子页面提取出next页面需要采集的url
/*
	1: 单进程任务可以通过管道来获取下一个页面需要采集和解析的内容
	2: 多进程任务可以通过broker.Add(item) 操作, 将item放入队列. 等待消费者获取
 */
func ReadYamlFileCreatedCrawler(){
	var wg scrapy.WaitGroupWrap
	var items = make(chan scrapy.ItemInterfaceI, 200)
	path := "producer.yaml"
	options, err := scrapy.NewOptions(path)
	if err != nil{
		panic(err)
	}
	for k, v := range options.Pages.Labels{
		wg.Add(1)
		go func(k string, v *scrapy.Page){
			defer wg.Done()
			var item = scrapy.NewMap()
			var parser = scrapy.NewGoQueryParser(v.Parser)
			item.Add(scrapy.NewPr("meta", v.Meta))
			scrapy.NewCrawler(v.Url, item).SetParser(parser).Do()
			items <- item
		}(k, v)
	}
	wg.Wait()
	close(items)

	for item := range items{
		log.Info(item)
	}
}

type Handler struct{}
func (h *Handler) HandleMessage(msg *nsq.Message) error{
	s := scrapy.String(msg.Body)
	log.Info(s.Hash())
	return nil
}

func ConsumerOptionsCreated(){
	go_utils.SetLogLevel("DEBUG")
	opt, err := scrapy.NewOptions("producer.yaml")
	if err != nil{
		panic(err)
	}
	/*
	consumer:
		通过nsq.Handler接口, 实现handler接收消费数据并进行处理的逻辑.
		default.Handler scrapy中有自己的默认配置, 建议在爬虫业务中, 契合自己的业务来实现对应的handler
	 */
	scrapy.NewRedisConsumer(opt.Consumer).SetHandler(&Handler{}).Run()
}

func GetHtml(){
	var item = scrapy.NewMap()
	var parser = scrapy.NewGoQueryParser("head title")
	var url scrapy.String = "https://www.toutiao.com/i6790992050591367684"
	var crawler = scrapy.NewCrawler(url, item).SetParser(parser).Do()
	log.Info(crawler.Html())
}



func main(){
	// 把采集结果放入redis队列中, 使用自带redis-broker方法
	BrokerDemo()

	// 通过读取yaml文件生成crawler需要的options
	ReadYamlFileCreatedCrawler()
	//ConsumerOptionsCreated()

	// 如果需要原始的html 可以通过以下方式来获取
	GetHtml()
}