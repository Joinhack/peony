package peony

import (
	"testing"
)

func TestRE(t *testing.T) {
	re := ruleRE
	arg := re.FindAllStringSubmatch("aaas", -1)
	t.Logf("%q\n", re.SubexpNames())
	t.Logf("%q\n", arg)
	arg = re.FindAllStringSubmatch("asdad<aw(sse):a><w>", -1)
	t.Logf("%q\n", re.SubexpNames())
	t.Logf("%q\n", arg)
}

func TestRouter(t *testing.T) {
	router := NewRouter()
	r := NewRoute("GET", "/path/12<string:oo>-<int:oo>", "action")
	err := router.complieRoute(r)
	if err != nil {
		panic(err)
	}
	println(r.PathRE.MatchString("/path/12-9090-123"))
}
