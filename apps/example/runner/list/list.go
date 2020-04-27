package list

import (
	"go-srapy/crawler"
	"go-srapy/internal"
)

var wg internal.WaitGroupWrap

func Main(options *crawler.Options){
	var ch = make(chan internal.ItemInterfaceI, 200)
	for _, page := range options.Pages.Labels{
		wg.Add(1)
		go func(page *crawler.Page){
			defer wg.Done()
			var item = internal.NewMap()
			item.Add(page.Meta)
			item.Add(internal.NewPr("next", page.Next))
			var parser = internal.NewGoQueryParser(page.Parser)
			var c = crawler.NewCrawler(page.Url, item)
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
