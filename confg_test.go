package peony

import (
	"strings"
	"testing"
)

func TestConfig1(t *testing.T) {
	content := `
key1=1
#comment
key2 = 2.2
key3 = y

=

	`
	lines := strings.Split(content, "\n")
	conf := Config{}
	if err := conf.readLines(lines); err != nil {
		t.Error(err)
	}
	t.Log(conf.Float("", "key2"))
	t.Log(conf.Int("", "key1"))
	t.Log(conf.Bool("", "key3"))
	t.Log(conf.String("", ""))
}

func TestConfig2(t *testing.T) {
	content := `
key1=1
#comment
key2 = 2.2
key3 = y

key4

	`
	lines := strings.Split(content, "\n")
	conf := Config{}
	if err := conf.readLines(lines); err != nil {
		t.Log(err)
	}

}

func TestConfig3(t *testing.T) {
	content := `
	aa=19
	[s1]
	key1=1
	#comment
	key2 = " 111"
	key3 = 10.2
	key3 = -$(key2)
	`
	lines := strings.Split(content, "\n")
	conf := Config{}
	if err := conf.readLines(lines); err != nil {
		t.Log(err)
	}
	t.Log(conf)
	t.Log(conf.Float("s1", "key2"))
	t.Log(conf.Int("s1", "key1"))
	t.Log(conf.Float("s1", "key3"))
	t.Log(conf.String("s1", "key3"))
	t.Log(conf.String("s1", "aa"))
}
