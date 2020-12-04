package scrapy

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/PuerkitoBio/goquery"
)

var log = go_utils.Log()

type Pattern map[string]interface{}

// 正则表达式解析
type R string

// goquery解析
type G string

//goquery解析html指定节点的text
type T string

//goquery解析html指定节点的attrib
type _A string

func A(pattern _A, attrib string) *GoQueryAttribParser {
	return &GoQueryAttribParser{
		pattern: String(pattern),
		attrib:  attrib,
	}
}

type GoQueryAttribParser struct {
	pattern String
	DefaultParser
	Result   *List
	attrib   string
	encoding string
}

func (g *GoQueryAttribParser) Encode(s string) IParser {
	g.encoding = s
	return g
}

func (g *GoQueryAttribParser) Validate() bool {
	return !g.Result.Empty()
}

func (g *GoQueryAttribParser) Parser(html String, item IItem, sss ...string) (IItem, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(AutoGetHtmlEncode(html.String(), g.encoding)))
	if err != nil {
		log.Error(err)
		return item, false
	}
	doc.Find(g.pattern.String()).Each(func(i int, selection *goquery.Selection) {
		if v, ok := selection.Attr(g.attrib); ok {
			if len(sss) == 0 {
				item.Add(NewPr(g.attrib, v))
			} else {
				item.Add(NewPr(sss[0], v))
			}
		}
	})
	return item, true
}

type ParserResult struct {
	Key   string      `bson:"key"`
	Value interface{} `bson:"value"`
}

func (p *ParserResult) String() string {
	return fmt.Sprintf(`(Key: %s), (Value: %v)`, p.Key, p.Value)
}

type IParser interface {
	Validate() bool
	Load(i IItem)
	Parser(string2 String, i IItem, string3 ...string) (IItem, bool)
	Encode(string) IParser
}

type DefaultParser struct {
	Result IItem
	sync.RWMutex
}

func (r *DefaultParser) Load(i IItem) {
	r.Lock()
	defer r.Unlock()
	r.Result = i
}

func (r *DefaultParser) Validate() bool {
	return !r.Result.Empty()
}

func NewRegexParser(pattern R) *RegexParser {
	return &RegexParser{
		Pattern: String(pattern),
	}
}

type GoQueryTextParser struct {
	Html    string
	pattern String
	DefaultParser
	encoding string
}

func (g *GoQueryTextParser) Encode(s string) IParser {
	g.encoding = s
	return g
}

func (g *GoQueryTextParser) Parser(html String, item IItem, sss ...string) (IItem, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(AutoGetHtmlEncode(html.String(), g.encoding)))
	if err != nil {
		return item, false
	}
	var texts []string
	doc.Find(g.pattern.String()).Each(func(i int, selection *goquery.Selection) {
		texts = append(texts, strings.TrimSpace(selection.Text()))
	})
	switch len(sss) > 0 {
	case true:
		item.Add(NewPr(sss[0], strings.Join(texts, ",")))
	default:
		item.Add(NewPr("text", strings.Join(texts, ",")))
	}
	return item, true
}

type GoQueryParser struct {
	Html    string
	pattern String
	DefaultParser
	encoding string
}

func (g *GoQueryParser) Encode(s string) IParser {
	g.encoding = s
	return g
}

func (g *GoQueryParser) Parser(html String, item IItem, sss ...string) (IItem, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(AutoGetHtmlEncode(html.String(), g.encoding)))
	if err != nil {
		log.Error(err)
		return item, false
	}
	var src = NewList()
	var href = NewList()
	doc.Find(g.pattern.String()).Each(func(i int, selection *goquery.Selection) {
		if s, ok := selection.Attr("src"); ok {
			src.Add(s)
		} else if h, ok := selection.Attr("href"); ok {
			href.Add(h)
		} else {
			key := newKey(html, sss...)
			h, _ := selection.Html()
			pr := NewPr(key, strings.TrimSpace(h))
			item.Add(pr)
		}
	})
	switch {
	case !src.Empty():
		item.Add(NewPr("src", src.Items()))
	case !href.Empty() && len(sss) != 0:
		item.Add(NewPr(sss[0], href.Items()))
	case !href.Empty():
		item.Add(NewPr("href", href.Items()))
	}
	return item, true
}

func NewGoQueryParser(pattern G) *GoQueryParser {
	return &GoQueryParser{
		pattern: String(pattern),
	}
}

func NewGoQueryTextParser(pattern T) *GoQueryTextParser {
	return &GoQueryTextParser{
		pattern: String(pattern),
	}
}

type JsonParser struct {
	DefaultParser
	Html     string
	encoding string
}

func (r *JsonParser) Encode(s string) IParser {
	r.encoding = s
	return r
}

func (r *JsonParser) Parser(htm String, interfaceI IItem, s ...string) (i IItem, ret bool) {
	var res = make(map[string]interface{})
	err := json.Unmarshal(htm.Decode(), &res)
	if err != nil {
		log.Warning("json parser unmarshal json error.")
		return
	}
	interfaceI.Add(res)
	ret = true
	return
}

func NewJsonParser() *JsonParser {
	return &JsonParser{}
}

type MixedParser struct {
	pattern Pattern
	DefaultParser
	encoding string
}

func (m *MixedParser) Encode(s string) IParser {
	m.encoding = s
	return m
}

func NewMixdParser(pattern Pattern) *MixedParser {
	return &MixedParser{
		pattern: pattern,
	}
}

func (m *MixedParser) Parser(html String, item IItem, s ...string) (i IItem, ret bool) {
	var res IParser
	for k, v := range m.pattern {
		switch t := v.(type) {
		case R:
			res = NewRegexParser(t)
		case G:
			res = NewGoQueryParser(t)
		case T:
			res = NewGoQueryTextParser(t)
		case *GoQueryAttribParser:
			res = t
		default:
			log.Debug(k)
			continue
		}
		res.Encode(m.encoding)
		res.Parser(html, item, k)
	}
	return
}

type RegexParser struct {
	Html    string
	Pattern String
	DefaultParser
	encoding string
}

func (r *RegexParser) Encode(s string) IParser {
	r.encoding = s
	return r
}

func (r *RegexParser) Parser(htm String, interfaceI IItem, s ...string) (i IItem, ret bool) {
	key := newKey(htm, s...)
	var result = Regex(AutoGetHtmlEncode(htm.String(), r.encoding), r.Pattern.String())
	if len(result) == 0 {
		return
	}
	interfaceI.Add(NewPr(key, result.Unique()))
	ret = true
	return
}

func newKey(html String, s ...string) string {
	if len(s) == 0 {
		return "scrapy.regex.parser"
	}
	return s[0]
}

func NewPr(key string, value interface{}) *ParserResult {
	return &ParserResult{
		Key:   key,
		Value: value,
	}
}
