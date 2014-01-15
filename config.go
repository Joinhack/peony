package peony

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type citem struct {
	Value string
	Next  *citem
}

var boolMapping = map[string]bool{
	"true":  true,
	"false": false,
	"yes":   true,
	"no":    false,
	"y":     true,
	"n":     false,
}

type Section map[string]*citem

type Config map[string]Section

func (s Section) HasItem(name string) bool {
	if _, ok := s[name]; ok {
		return true
	}
	return false
}

func (c Config) HasSection(name string) bool {
	if _, ok := c[name]; ok {
		return true
	}
	return false
}

func (conf Config) readLines(lines []string) (err error) {
	var secname string
	var section Section
	for l, line := range lines {
		var idx int
		line = strings.Trim(line, " \t\n")
		if len(line) == 0 || line[0] == '#' {
			//skip empty line or coment line
			continue
		}
		lineLen := len(line)
		if len(line) >= 2 && line[0] == '[' && line[lineLen-1] == ']' {
			secname = line[1 : lineLen-1]
			section = nil
			continue
		}
		if section == nil {
			if conf.HasSection(secname) {
				section = conf[secname]
			} else {
				section = make(Section)
				conf[secname] = section
			}
		}

		if idx = strings.Index(line, "="); idx == -1 {
			err = errors.New(fmt.Sprintf("key value expr error, @line: %d", l))
			return
		}
		key := strings.Trim(line[0:idx], " \t\n")
		val := strings.Trim(line[idx+1:], " \t\n")
		item := &citem{val, nil}
		if section.HasItem(key) {
			item.Next = section[key]
		}
		section[key] = item
	}
	return
}

func (conf Config) ReadFile(path string) (err error) {
	var lines []string
	lines, err = ReadLines(path)
	if err != nil {
		return
	}
	err = conf.readLines(lines)
	return
}

func (c Config) getItem(secname, key string) *citem {
	var section Section
	var item *citem
	var ok bool

	secs := make([]Section, 2)

	if section, ok = c[secname]; ok {
		secs = append(secs, section)
	}
	if section, ok = c[""]; ok {
		secs = append(secs, section)
	}
	for _, section := range secs {
		if item, ok = section[key]; ok {
			return item
		}
	}
	return nil
}

func (c Config) Bool(secname, key string) (rs, ok bool) {
	var item *citem
	ok = false
	for item = c.getItem(secname, key); item != nil; item = item.Next {
		val := strings.ToLower(c.value(secname, item.Value))
		rs, ok = boolMapping[val]
		if !ok {
			continue
		}
		ok = true
		return
	}
	return
}

func (c Config) IntDefault(secname, key string, i int64) int64 {
	var val int64
	var ok bool
	if val, ok = c.Int(secname, key); ok {
		return val
	}
	return i
}

func (c Config) FloatDefault(secname, key string, f float64) float64 {
	var val float64
	var ok bool
	if val, ok = c.Float(secname, key); ok {
		return val
	}
	return f
}

func (c Config) StringDefault(secname, key, s string) string {
	var val string
	var ok bool
	if val, ok = c.String(secname, key); ok {
		return val
	}
	return s
}

func (c Config) Int(secname, key string) (rs int64, ok bool) {
	var item *citem
	ok = false
	for item = c.getItem(secname, key); item != nil; item = item.Next {
		val := c.value(secname, item.Value)
		var err error
		if strings.HasPrefix(val, "0x") || strings.HasPrefix(val, "0X") {
			if rs, err = strconv.ParseInt(val, 16, 64); err != nil {
				continue
			}
			ok = true
			return
		}

		if strings.HasPrefix(item.Value, "o") || strings.HasPrefix(item.Value, "O") {
			if rs, err = strconv.ParseInt(val, 8, 64); err != nil {
				continue
			}
			ok = true
			return
		}
		if rs, err = strconv.ParseInt(val, 10, 64); err != nil {
			continue
		}
		ok = true
		return
	}
	return
}

func (c Config) Float(secname, key string) (rs float64, ok bool) {
	var item *citem
	ok = false
	var err error
	for item = c.getItem(secname, key); item != nil; item = item.Next {
		if rs, err = strconv.ParseFloat(c.value(secname, item.Value), 64); err != nil {
			continue
		}
		ok = true
		return
	}
	return
}

var (
	varRE = regexp.MustCompile(`\$\(([^)]*)\)?`)
)

func (c Config) value(secname, val string) string {
	matched := varRE.FindAllStringIndex(val, -1)
	idx := 0
	slice := []string{}
	var value string
	var ok bool
	for _, match := range matched {
		if idx < match[0] {
			slice = append(slice, val[idx:match[0]])
		}
		expr := val[match[0]:match[1]]

		if value, ok = c.String(secname, expr[2:len(expr)-1]); !ok {
			value = ""
		}
		slice = append(slice, value)
		idx = match[1]
	}
	if idx < len(val) {
		slice = append(slice, val[idx:len(val)])
	}
	return strings.Join(slice, "")
}

func (c Config) String(secname, key string) (rs string, ok bool) {
	var item *citem
	if item = c.getItem(secname, key); item == nil {
		return
	}
	ok = true
	rs = item.Value
	if len(rs) >= 2 && rs[0] == '"' && rs[len(rs)-1] == '"' {
		rs = rs[1 : len(rs)-1]
		return
	}
	rs = c.value(secname, rs)
	return
}
