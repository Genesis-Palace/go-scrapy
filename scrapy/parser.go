package scrapy

import (
	"encoding/json"
	"fmt"
	go_utils "github.com/Genesis-Palace/go-utils"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"strings"
	"sync"
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

func A(pattern _A, attrib ...string) *GoQueryAttribParser {
	if len(attrib) == 0{
		return &GoQueryAttribParser{
			pattern: String(pattern),
		}
	}
	return &GoQueryAttribParser{
		pattern: String(pattern),
		attrib: String(attrib[0]),
	}
}

type GoQueryAttribParser struct {
	pattern String
	DefaultParser
	Result *List
	attrib String
}

func (g *GoQueryAttribParser) Validate() bool{
	return !g.Result.Empty()
}

func (g *GoQueryAttribParser) Parser(html String, item ItemInterfaceI, sss ...string) (ItemInterfaceI, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(AutoGetHtmlEncode(html.String())))
	if err != nil {
		log.Error(err)
		return item, false
	}
	attrib := NewList()
	doc.Find(g.pattern.String()).Each(func(i int, selection *goquery.Selection) {
		if v, ok := selection.Attr(g.attrib.String()); ok {
			attrib.Add(v)
		}
	})
	if !g.attrib.Empty(){
		item.Add(NewPr(g.attrib.String(), attrib.Items()))
	}else {
		item.Add(NewPr("attribs", attrib.Items()))
	}
	return item, true
}

type ParserResult struct {
	Key   string
	Value interface{}
}

func (p *ParserResult) String() string {
	return fmt.Sprintf(`(Key: %s), (Value: %v)`, p.Key, p.Value)
}

type ParserInterfaceI interface {
	Validate() bool
	Load(i ItemInterfaceI)
	Parser(string2 String, i ItemInterfaceI, string3 ...string) (ItemInterfaceI, bool)
}

type DefaultParser struct {
	Result ItemInterfaceI
	sync.RWMutex
}

func (r *DefaultParser) Load(i ItemInterfaceI) {
	r.Lock()
	defer r.Unlock()
	i = r.Result
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
}

func (g *GoQueryTextParser) Parser(html String, item ItemInterfaceI, sss ...string) (ItemInterfaceI, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(AutoGetHtmlEncode(html.String())))
	if err != nil {
		log.Error(err)
		return item, false
	}
	texts := NewList()
	doc.Find(g.pattern.String()).Each(func(i int, selection *goquery.Selection) {
		texts.Add(strings.TrimSpace(selection.Text()))
	})
	switch len(sss) > 0 {
	case true:
		item.Add(NewPr(sss[0], texts.Items()))
	default:
		item.Add(NewPr("text", texts.Items()))
	}
	return item, true
}

type GoQueryParser struct {
	Html    string
	pattern String
	DefaultParser
}

func (g *GoQueryParser) Parser(html String, item ItemInterfaceI, sss ...string) (ItemInterfaceI, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(AutoGetHtmlEncode(html.String())))
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
	Html    string
	pattern String
}

func (r *JsonParser) Parser(htm String, interfaceI ItemInterfaceI, s ...string) (i ItemInterfaceI, ret bool) {
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
}

func NewMixdParser(pattern Pattern) *MixedParser {
	return &MixedParser{
		pattern: pattern,
	}
}

func (m *MixedParser) Parser(html String, item ItemInterfaceI, s ...string) (i ItemInterfaceI, ret bool) {
	var res ParserInterfaceI
	for k, v := range m.pattern {
		switch v.(type) {
		case R:
			res = NewRegexParser(v.(R))
		case G:
			res = NewGoQueryParser(v.(G))
		case T:
			res = NewGoQueryTextParser(v.(T))
		case *GoQueryAttribParser:
			res = v.(*GoQueryAttribParser)
		default:
			log.Debug(k)
			continue
		}
		res.Parser(html, item, k)
	}
	return
}

type RegexParser struct {
	Html    string
	Pattern String
	DefaultParser
}

func (r *RegexParser) Parser(htm String, interfaceI ItemInterfaceI, s ...string) (i ItemInterfaceI, ret bool) {
	key := newKey(htm, s...)
	var result = Regex(AutoGetHtmlEncode(htm.String()), r.Pattern.String())
	if len(result) == 0 {
		return
	}
	pr := NewPr(key, html.UnescapeString((result[0])))
	interfaceI.Add(pr)
	ret = true
	return
}

func newKey(html String, s ...string) string {
	if len(s) == 0 {
		return html.Hash()
	}
	return s[0]
}

func NewPr(key string, value interface{}) *ParserResult {
	return &ParserResult{
		Key:   key,
		Value: value,
	}
}
