package scrapy

func DefaultPipelines(i IItem) {
	if d, e := i.Dumps(); e == nil {
		log.Debug(d)
	}
}
