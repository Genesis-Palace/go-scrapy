package list

import (
	"go-srapy/crawler"
	"go-srapy/internal"
)

// 目前代理只支持阿布云代理类型
// 使用方式已经提供, 如下, 创建一个新的阿布云代理client即可
func MainProxyList(options *crawler.Options){
	var ch = make(chan internal.ItemInterfaceI, 200)
	for _, page := range options.Pages.Labels{
		wg.Add(1)
		go func(page *crawler.Page){
			defer wg.Done()
			var item = internal.NewMap()
			var proxy = internal.NewAbutunProxy(
				"H01234567890123P",
				"0123456789012345",
				"http-pro.abuyun.com:9010",
				)
			item.Add(page.Meta)
			item.Add(internal.NewPr("next", page.Next))
			var parser = internal.NewGoQueryParser(page.Parser)
			var c = crawler.NewProxyCrawler(page.Url, proxy, item)
			c.SetTimeOut(3).SetParser(parser).Do()
			ch <- item
		}(page)
	}
	wg.Wait()
	close(ch)

	for v := range ch{
		options.Broker.Add(v)
	}
}