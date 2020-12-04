package scrapy

import (
	"github.com/axgle/mahonia"
	"github.com/elliotchance/pie/pie"
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

type RegexItems []RegexItem

type RegexItem struct {
	key interface{}
	val interface{}
}

func (r *RegexItem) StringVal() string {
	v, ok := r.val.(string)
	if ok {
		return v
	}
	return ""
}

func (r RegexItems) Val() pie.Strings {
	var result = pie.Strings{}
	for _, item := range r {
		result = result.Append(item.StringVal())
	}
	return result
}

func (r RegexItems) First() string {
	return r[0].StringVal()
}

// RegexParse : 通过正则表达式提取 html中的指定 regex 元素
func Regex(html, rex string) pie.Strings {
	regex := regexp.MustCompile(rex)
	find := regex.FindAllStringSubmatch(html, -1)
	if len(find) == 0 || len(find[0]) <= 1 {
		return pie.Strings{}
	}
	var regexItems = RegexItems{}
	for _, item := range find {
		regexItems = append(regexItems, RegexItem{item[0], item[1]})
	}
	return regexItems.Val()
}

func AutoGetHtmlEncode(html string, encode string) string {
	if encode == "" {
		encode = "utf-8"
	}
	dec := mahonia.NewDecoder(encode)
	return dec.ConvertString(html)
}
