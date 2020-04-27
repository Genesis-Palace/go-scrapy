package main

import (
	"flag"
	go_utils "github.com/Genesis-Palace/go-utils"
	"go-srapy/apps/example/runner/detail"
	"go-srapy/apps/example/runner/list"
	"go-srapy/crawler"
)

var (
	cfg    string
	broker string
	proxy  bool
	isL    bool
	isC    bool
)

func init() {
	flag.StringVar(&cfg, "c", "apps/config/test_list.yaml", "采集程序配置文件")
	flag.BoolVar(&isL, "list", false, "是否为列表采集程序. 默认为true")
	flag.BoolVar(&proxy, "p", false, "是否为列表采集程序. 默认为true")
	flag.BoolVar(&isC, "consumer", true, "consumer 爬虫消费分布节点")
	flag.StringVar(&broker, "broker", "apps/config/broker.yaml", "采集程序配置文件")
	flag.Parse()
}

func init() {
	go_utils.SetLogLevel("DEBUG")
}

func main() {
	options, err := crawler.NewOptions(cfg)
	if err != nil {
		panic(err)
	}
	switch {
	case isL:
		if proxy{
			list.MainProxyList(options)
		}else{
			list.Main(options)
		}
	case isC:
		switch {
		case options.Consumer.Redis != nil:
			detail.RedisQueueMain(options.Consumer)
		case options.Consumer.Nsq != nil:
			detail.NsqQueueMain(options.Consumer)
		}
	}
}
