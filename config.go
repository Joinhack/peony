package peony

import (
	"errors"
	"fmt"
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

	if section, ok = c[secname]; !ok {
		return nil
	}
	if item, ok = section[key]; ok {
		return item
	}
	return nil
}

func (c Config) Bool(secname, key string) (rs, ok bool) {
	var item *citem
	ok = false
	for item = c.getItem(secname, key); item != nil; item = item.Next {
		val := strings.ToLower(item.Value)
		rs, ok = boolMapping[val]
		if !ok {
			continue
		}
		ok = true
		return
	}
	return
}

func (c Config) Int(secname, key string) (rs int64, ok bool) {
	var item *citem
	ok = false
	for item = c.getItem(secname, key); item != nil; item = item.Next {
		var err error
		if strings.HasPrefix(item.Value, "0x") || strings.HasPrefix(item.Value, "0X") {
			if rs, err = strconv.ParseInt(item.Value, 16, 64); err != nil {
				continue
			}
			ok = true
			return
		}

		if strings.HasPrefix(item.Value, "o") || strings.HasPrefix(item.Value, "O") {
			if rs, err = strconv.ParseInt(item.Value, 8, 64); err != nil {
				continue
			}
			ok = true
			return
		}
		if rs, err = strconv.ParseInt(item.Value, 10, 64); err != nil {
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
		if rs, err = strconv.ParseFloat(item.Value, 64); err != nil {
			continue
		}
		ok = true
		return
	}
	return
}

func (c Config) String(secname, key string) (rs string, ok bool) {
	var item *citem
	if item = c.getItem(secname, key); item == nil {
		return
	}
	ok = true
	rs = item.Value
	return
}
