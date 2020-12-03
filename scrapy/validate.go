package scrapy

import (
	"github.com/axgle/mahonia"
	"regexp"
	"sync"

	"gopkg.in/go-playground/validator.v9"
)

var once sync.Once

func Validated(s interface{}) bool {
	v := validator.New()
	err := v.Struct(s)
	return err == nil
}

// RegexParse : 通过正则表达式提取 html中的指定 regex 元素
func Regex(html, rex string) [][]string {
	regex := regexp.MustCompile(rex)
	find := regex.FindAllStringSubmatch(html, -1)
	if len(find) == 0 || len(find[0]) <= 1 {
		return []string{}
	}
	return find
}

func AutoGetHtmlEncode(html string, encode string) string {
	if encode == "" {
		encode = "utf-8"
	}
	dec := mahonia.NewDecoder(encode)
	return dec.ConvertString(html)
}
