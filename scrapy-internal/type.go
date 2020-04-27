package scrapy_internal

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
	"sync"
)

type String string

func (s *String) Replace(pattern string) String {
	return String(strings.ReplaceAll(string(*s), pattern, ""))
}

func (s *String) Decode() []byte {
	return []byte(*s)
}

func (s *String) String() string {
	return string(*s)
}

func (s *String) Hash() string {
	h := md5.New()
	h.Write([]byte(*s))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *String) HasPrefix(pattern string) bool {
	return strings.HasPrefix(string(*s), pattern)
}

func (s *String) Empty() bool {
	return *s == ""
}

type Map struct {
	sync.RWMutex
	m map[String]interface{}
}

func (m *Map) Contains(s string) bool {
	_, ok := m.m[String(s)]
	return ok
}
func (m *Map) Items() map[String]interface{} {
	return m.m
}

func (m *Map) Load(b []byte) error {
	err := json.Unmarshal(b, &m.m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Map) Pop(s String) interface{} {
	m.Lock()
	m.Unlock()
	return m.del(s)
}

func (m *Map) Add(v interface{}) {
	m.Lock()
	switch v.(type) {
	case map[string]string:
		for key, value := range v.(map[string]string) {
			m.m[String(key)] = value
		}
	case map[string]interface{}:
		for key, value := range v.(map[string]interface{}) {
			m.m[String(key)] = value
		}
	case *ParserResult:
		m.m[String(v.(*ParserResult).Key)] = v.(*ParserResult).Value
	}
	m.Unlock()
}

func (m *Map) Size() int {
	return len(m.m)
}

func (m *Map) Dumps() (String, error) {
	js, err := json.Marshal(m.m)
	if err != nil {
		return "", err
	}
	return String(js), nil
}

func (m *Map) del(k String) interface{} {
	if v, ok := m.m[k]; ok {
		delete(m.m, k)
		return v
	}
	return nil
}

func (m *Map) Get(k String) interface{} {
	m.RLock()
	v := m.m[k]
	m.RUnlock()
	return v
}

func (m *Map) Empty() bool {
	return m.Size() == 0
}

type List struct {
	sync.RWMutex
	l []interface{}
}

func (l *List) Contains(s string) bool {
	return true
}

func (l *List) Add(item interface{}) {
	l.Lock()
	defer l.Unlock()
	l.l = append(l.l, item)
}

func (l *List) Size() int {
	return len(l.l)
}

func (l *List) Dumps() (String, error) {
	js, err := json.Marshal(l.l)
	if err != nil {
		return "", err
	}
	return String(js), nil
}

func (l *List) Items() []interface{} {
	return l.l
}

func (l *List) Empty() bool {
	return l.Size() == 0
}

func NewMap() *Map {
	return &Map{
		RWMutex: sync.RWMutex{},
		m:       make(map[String]interface{}),
	}
}

func NewList() *List {
	return &List{
		RWMutex: sync.RWMutex{},
		l:       []interface{}{},
	}
}
