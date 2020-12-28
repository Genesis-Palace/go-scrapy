package scrapy

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/elliotchance/pie/pie"
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
	_, err := h.Write([]byte(*s))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (s *String) HasPrefix(pattern string) bool {
	return strings.HasPrefix(string(*s), pattern)
}

func (s *String) Empty() bool {
	return strings.TrimSpace(s.String()) == ""
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
	defer m.Unlock()
	return m.del(s)
}

func (m *Map) Add(v interface{}) {
	m.Lock()
	switch t := v.(type) {
	case *Map:
		for key, value := range t.Items() {
			m.m[key] = value
		}
	case map[string]string:
		for key, value := range t {
			m.m[String(key)] = value
		}
	case map[string]interface{}:
		for key, value := range t {
			m.m[String(key)] = value
		}
	case *ParserResult:
		m.m[String(t.Key)] = t.Value
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

type StringList struct {
	sync.RWMutex
	l pie.Strings
}

func (l *StringList) Contains(s string) bool {
	return true
}

func (l *StringList) Add(item interface{}) {
	l.Lock()
	defer l.Unlock()
	l.l = l.l.Append(item.(string))
}

func (l *StringList) Size() int {
	return l.l.Len()
}

func (l *StringList) Dumps() (String, error) {
	js, err := json.Marshal(l.l)
	if err != nil {
		return "", err
	}
	return String(js), nil
}

func (l *StringList) Items() pie.Strings {
	return l.l
}

func (l *StringList) Load(b []byte) error {
	err := json.Unmarshal(b, &l.l)
	if err != nil {
		return err
	}
	return nil
}

func (l *StringList) Empty() bool {
	return l.Size() == 0
}

func NewMap() *Map {
	return &Map{
		RWMutex: sync.RWMutex{},
		m:       make(map[String]interface{}),
	}
}

func NewStringList() *StringList {
	return &StringList{
		RWMutex: sync.RWMutex{},
	}
}
