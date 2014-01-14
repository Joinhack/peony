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

func ConfigRead(path string) (conf Config, err error) {
	var lines []string
	lines, err = ReadLines(path)
	if err != nil {
		return
	}
	var secname string
	var section Section
	conf = make(Config)
	for l, line := range lines {
		var idx int
		line = strings.Trim(line, " \t\n")
		lineLen := len(line)
		if len(line) >= 2 && line[0] == '[' && line[lineLen-1] == ']' {
			secname = line[1 : lineLen-2]
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
			err = errors.New(fmt.Sprintf("unkown key value expr, line:", l))
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
	if item = c.getItem(secname, key); item == nil {
		ok = false
		return
	}
	val := strings.ToLower(item.Value)
	rs, ok = boolMapping[val]
	return
}

func (c Config) Int(secname, key string) (rs int64, ok bool) {
	var item *citem
	if item = c.getItem(secname, key); item == nil {
		ok = false
		return
	}
	var err error
	if rs, err = strconv.ParseInt(item.Value, 10, 64); err != nil {
		ok = false
	}
	ok = true
	return
}

func (c Config) String(secname, key string) (rs string, ok bool) {
	var item *citem
	if item = c.getItem(secname, key); item == nil {
		ok = false
		return
	}
	ok = true
	rs = item.Value
	return
}
