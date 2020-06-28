package main

import (
	"errors"
	"fmt"
	"github.com/Genesis-Palace/go-scrapy/scrapy"
	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/nsqio/go-nsq"
)

var (
	log    = go_utils.Log()
	broker = scrapy.NewRedisBroker("127.0.0.1:6379", "", "test-redis-broker", 0)
)

// 代理设置方法
func example3() {
	// 目前代理只支持阿布云. 没有提供client的相关接口
	var item = scrapy.NewMap()
	var proxy = scrapy.NewAbutunProxy("appid", "secret", "proxyserver")
	var parser = scrapy.NewGoQueryParser("head title")
	var url scrapy.String = "https://www.toutiao.com/i6790992050591367684"
	scrapy.NewProxyCrawler(url, proxy, item).SetParser(parser).SetTimeOut(1).Do()
	fmt.Println(item.Items())
}

func BrokerDemo() {
	var url scrapy.String = "https://www.toutiao.com/i6820706041991266827"
	var item = scrapy.NewMap()
	var parser = scrapy.NewMixdParser(scrapy.Pattern{
		"title":    scrapy.G("head title"),
		"abstract": scrapy.R("abstract: '(.*?)'"),
	})
	scrapy.NewCrawler(url, item).SetTimeOut(5).SetParser(parser).Do()
	if item.Empty() {
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
func ReadYamlFileCreatedCrawler() {
	var wg scrapy.WaitGroupWrap
	var items = make(chan scrapy.IItem, 200)
	path := "producer.yaml"
	options, err := scrapy.NewOptions(path)
	if err != nil {
		panic(err)
	}
	for k, v := range options.Pages.Labels {
		wg.Add(1)
		go func(k string, v *scrapy.Page) {
			defer wg.Done()
			var item = scrapy.NewMap()
			var parser = scrapy.NewGoQueryParser(v.Parser)
			item.Add(scrapy.NewPr("meta", v.Meta))
			item.Add(scrapy.NewPr("next", v.Next))
			scrapy.NewCrawler(v.Url, item).SetParser(parser).Do()
			items <- item
			broker.Add(item)
		}(k, v)
	}
	wg.Wait()
	close(items)

	for item := range items {
		log.Info(item)
	}
}

// Handler: redisConsumer message handler
type Handler struct{}

func (h *Handler) HandleMessage(msg *nsq.Message) error {
	var item = scrapy.NewMap()
	err := item.Load(msg.Body)
	if err != nil {
		return err
	}
	var nextParser scrapy.Pattern
	if v := item.Get("next"); v != nil {
		next, err := scrapy.NewNext(v)
		if err != nil {
			return err
		}
		nextParser = next.MergeGr()
	}
	var hrefs = item.Get("href")
	if hrefs == nil {
		return errors.New("next hrefs is empty")
	}
	// 创建rss item
	var feeds = scrapy.NewFeeds()
	for _, url := range hrefs.([]interface{}) {
		var detail = scrapy.NewMap()
		scrapy.NewCrawler(scrapy.String(url.(string)), detail).SetParser(scrapy.NewMixdParser(nextParser)).Do()
		// 把detail放入feeds结构中
		feeds.Add(detail)
	}
	// feeds dumps落地当前结构中的生成的xml, 提供给service使用
	s, err := feeds.Dumps()
	if err != nil {
		return err
	}
	log.Info(s.String())
	return nil
}

func ConsumerOptionsCreated() {
	go_utils.SetLogLevel("DEBUG")
	opt, err := scrapy.NewOptions("producer.yaml")
	if err != nil {
		panic(err)
	}
	/*
		consumer:
			通过nsq.Handler接口, 实现handler接收消费数据并进行处理的逻辑.
			default.Handler scrapy中有自己的默认配置, 建议在爬虫业务中, 契合自己的业务来实现对应的handler
	*/
	scrapy.NewRedisConsumer(opt.Consumer).SetHandler(&Handler{}).Run()
}

func GetHtml() {
	var item = scrapy.NewMap()
	var parser = scrapy.NewGoQueryParser("head title")
	var url scrapy.String = "https://www.toutiao.com/i6790992050591367684"
	var crawler, _ = scrapy.NewCrawler(url, item).SetParser(parser).Do()
	log.Info(crawler.Html())
}

func main() {
	//把采集结果放入redis队列中, 使用自带redis-broker方法
	//var wg scrapy.WaitGroupWrap
	//wg.Wrap(func(){
	//	for i:=0; i<=5; i++{
	//		go func(){
	//			BrokerDemo()
	//		}()
	//		time.Sleep(time.Second)
	//	}
	//})
	//
	////通过读取yaml文件生成crawler需要的options
	//ReadYamlFileCreatedCrawler()
	////读取 ReadYamlFileCreatedCrawler 写入管道中的信息, 通过HandlerMessage方法进行消息处理.
	////实现多端分布式逻辑
	//wg.Wrap(ConsumerOptionsCreated)
	//
	////如果需要原始的html 可以通过以下方式来获取
	//GetHtml()
	//wg.Wait()
	//// 增加jsonparser基础实现, 将response.Html转成map 并写入item中. 解析由开发者自定义即可
	//NewToutiaoCrawlerJsonParser()
	//NewSoHuNewsJSONParser()
	NewKafkaConsumerTest()
}

func NewToutiaoCrawlerJsonParser() {
	// 打开debug log
	go_utils.SetLogLevel("DEBUG")
	var url scrapy.String = "https://www.toutiao.com/article/v2/tab_comments/?aid=24&app_name=toutiao-web&group_id=6821094670118945293&item_id=6821094670118945293&offset=0&count=5"
	var item = scrapy.NewMap()
	c, _ := scrapy.NewCrawler(url, item).SetParser(scrapy.NewJsonParser()).Do()
	c.Html()
	log.Debug(item)
}

func NewSoHuNewsJSONParser() {
	//go_utils.SetLogLevel("DEBUG")
	var url scrapy.String = "https://apiv2.sohu.com/api/topic/load?callback=&page_size=10&topic_source_id=mp_392096309&page_no=1&hot_size=5&media_id=267106&topic_category_id=8&topic_title=%E4%B8%AD%E5%85%B1%E4%B8%AD%E5%A4%AE%E6%94%BF%E6%B2%BB%E5%B1%80%E5%B8%B8%E5%8A%A1%E5%A7%94%E5%91%98%E4%BC%9A%E5%8F%AC%E5%BC%80%E4%BC%9A%E8%AE%AE%E4%B9%A0%E8%BF%91%E5%B9%B3%E4%B8%BB%E6%8C%81&topic_url=https%3A%2F%2Fwww.sohu.com%2Fa%2F392096309_267106%3Fcode%3D59d2c479cc76988d098d8b3251ed61c4&source_id=mp_392096309&_=1588211540064"
	var item = scrapy.NewMap()
	_, _ = scrapy.NewCrawler(url, item).SetTimeOut(5).SetParser(scrapy.NewJsonParser()).Do()
	for key, value := range item.Get("jsonObject").(map[string]interface{}) {
		log.Info(key, value)
	}
}

type KafkaHandler struct{}

func (k KafkaHandler) HandleMessage(message *nsq.Message) error {
	log.Info(scrapy.String(message.Body))
	return nil
}

func NewKafkaConsumerTest() {
	var wg scrapy.WaitGroupWrap
	wg.Add(1)
	go func() {
		defer wg.Done()
		opt, err := scrapy.NewOptions("producer.yaml")
		if err != nil {
			panic(err)
		}
		scrapy.NewKafkaConsumer(opt.Consumer).SetHandler(KafkaHandler{}).Run()
	}()
	broker := scrapy.NewKafkaBroker([]string{"127.0.0.1:9092"}, "test")
	var url scrapy.String = "https://apiv2.sohu.com/api/topic/load?callback=&page_size=10&topic_source_id=mp_392096309&page_no=1&hot_size=5&media_id=267106&topic_category_id=8&topic_title=%E4%B8%AD%E5%85%B1%E4%B8%AD%E5%A4%AE%E6%94%BF%E6%B2%BB%E5%B1%80%E5%B8%B8%E5%8A%A1%E5%A7%94%E5%91%98%E4%BC%9A%E5%8F%AC%E5%BC%80%E4%BC%9A%E8%AE%AE%E4%B9%A0%E8%BF%91%E5%B9%B3%E4%B8%BB%E6%8C%81&topic_url=https%3A%2F%2Fwww.sohu.com%2Fa%2F392096309_267106%3Fcode%3D59d2c479cc76988d098d8b3251ed61c4&source_id=mp_392096309&_=1588211540064"
	var item = scrapy.NewMap()
	_, _ = scrapy.NewCrawler(url, item).SetTimeOut(5).SetParser(scrapy.NewJsonParser()).Do()
	broker.Add(item)
	wg.Wait()
}
