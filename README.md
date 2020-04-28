#go-scrapy开发文档
@(工具开发, 效率, 爬虫)

-------
###项目结构
- go-scrapy
	- scrapy-internal
		- ItemInterface
		- ParserInterface
		- ClientInterface
		- autoHtmlEncode
		- waitGroupWrap
		- types
			- Map
			- List
			- Url
			- String
	- scrapy
		- Options
		- BrokerInterface
			- nsqBroker
			- redisBroker
		- ConsumerInterface
			- nsqConsumer
			- redisConsumer 

-----
>设计思路

	- 不想每次都用大量重复的代码开发爬虫, 抽取爬虫的核心思路, 拆成几个模块
		- request
		- pipelines
		- proxy
		- 分布式
		- 配置管理
	- 通过简单的yaml配置文件, 即可快速运行爬虫, 实现分布式采集等相关任务
	- 主要是爬虫任务太多的时候, 写烦了 [○･｀Д´･ ○].

----

>可扩展性

	- 面向接口编程的理念. 解决了部分瑞士军刀代码的问题
	- 面向接口的开发可以使开发者在使用go-scrapy的同时, 便于扩展

---

```go
import (
	"fmt"
	"github.com/Genesis-Palace/go-scrapy/scrapy"
	internal "github.com/Genesis-Palace/go-scrapy/scrapy-internal"
	go_utils "github.com/Genesis-Palace/go-utils"
	"sync"
)

var (
	log = go_utils.Log()
)

func main(){
	var url internal.String
	url = "https://www.toutiao.com/i6790992050591367684"
	// go-scrapy内定义了多种item类型, 如需扩展, 完成接口定义即可
	// 详见 scrapy-internal中的ItemInterfaceI接口
	var item = internal.NewMap()
	scrapy.NewCrawler(url, item).SetParser(internal.NewMixdParser(internal.Pattern{
		"title": internal.G("head title"),
		"source": internal.R(`source: '(.*?)'`),
		"abstract": internal.R(`abstract: '(.*?)'`),
	})).SetTimeOut(1).Do()
	/* 
		parser支持3种方式, goquery解析, regex正则解析以及goquery和正则混合解析的方式.
		* internal.G 使用goquery解析
			* type: string
		* internal.R 使用正则解析
			* type string
		* internal.Pattern 混合类型
			* type map[string]interface{}
		parser提供接口, 如需扩展, 完成接口实现即可.
	*/
	log.Info(item.Items())
}
```

```go
// 代理设置方法
func example3(){
	// 目前代理只支持阿布云. 没有提供client的相关接口
	var item = internal.NewMap()
	var proxy = internal.NewAbutunProxy("appid", "secret", "proxyserver")
	var parser = internal.NewGoQueryParser("head title")
	var url internal.String = "https://www.toutiao.com/i6790992050591367684"
	scrapy.NewProxyCrawler(url, proxy, item).SetParser(parser).SetTimeOut(1).Do()
	fmt.Println(item.Items())
```

如果对项目有兴趣的同学可以试用, 遇到问题issue一下,  希望大家喜欢这个小工具(*^▽^*)