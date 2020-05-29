package scrapy

import (
	"regexp"
	"strings"
	"sync"

	"github.com/axgle/mahonia"
	"gopkg.in/go-playground/validator.v9"
)

var once sync.Once

func Validated(s interface{}) bool {
	v := validator.New()
	err := v.Struct(s)
	return err == nil
}

// RegexParse : 通过正则表达式提取 html中的指定 regex 元素
func Regex(html, rex string) []string {
	regex := regexp.MustCompile(rex)
	find := regex.FindAllStringSubmatch(html, -1)
	if len(find) == 0 || len(find[0]) <= 1 {
		return []string{}
	}
	return []string{find[0][1]}
}

func AutoGetHtmlEncode(html string) string {
	enCode := Regex(html, "<meta.*?charset=(.*?)>")
	if len(enCode) == 0 || enCode[0] == "" {
		return ""
	}
	code := enCode[0]
	if strings.Contains(code, "utf-8") || strings.Contains(code, "UTF-8") {
		code = "UTF-8"
	} else if strings.Contains(code, "2312") {
		code = "GB18030"
	} else if strings.Contains(code, "gbk") || strings.Contains(code, "GBK") {
		code = "GBK"
	} else {
		code = ""
	}
	dec := mahonia.NewDecoder(code)
	return dec.ConvertString(html)
}
