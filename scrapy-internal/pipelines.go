package scrapy_internal

func DefaultPipelines(i ItemInterfaceI) {
	if d, e := i.Dumps(); e == nil {
		log.Debug(d)
	}
}
