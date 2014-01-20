package peony

import (
	"errors"
	tmpl "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	NotDirError     = errors.New("Is not a directory")
	NoTemplateFound = errors.New("no template found")
)

type TemplateLoader struct {
	template       *tmpl.Template //
	basePath       []string
	FuncMap        tmpl.FuncMap
	extendParams   map[string]interface{}
	loadedTemplate map[string]string //key is template name, value is path
	forceNotify    bool              //force notify when the first time
}

func (t *TemplateLoader) IgnoreDir(file os.FileInfo) bool {
	if strings.HasPrefix(file.Name(), ".") {
		return true
	}
	return false
}

//when the suffix is .html notify the change
func (t *TemplateLoader) IgnoreFile(file string) bool {
	switch filepath.Ext(file) {
	case "html", "json", "xml":
	default:
		return false
	}
	return true
}

func (t *TemplateLoader) Refresh() error {
	t.reset()
	return t.load()
}

func (t *TemplateLoader) ForceRefresh() bool {
	return t.forceNotify
}

func (t *TemplateLoader) Path() []string {
	return t.basePath
}

func (t *TemplateLoader) skipDir(n string) bool {
	return strings.HasPrefix(n, ".")
}

func (t *TemplateLoader) skipFile(n string) bool {
	return strings.HasPrefix(n, ".")
}

func templateName(path string) string {
	return filepath.ToSlash(path)
}

func NewTemplateLoader(base []string) *TemplateLoader {
	tl := &TemplateLoader{
		basePath:     base,
		forceNotify:  true,
		FuncMap:      tmpl.FuncMap{},
		extendParams: map[string]interface{}{},
	}
	tl.FuncMap["ExtParams"] = func() map[string]interface{} { return tl.extendParams }
	return tl
}

func (t *TemplateLoader) BindServerTemplateFunc(svr *Server) {
	t.FuncMap["IsDevMode"] = func() bool {
		return svr.App.DevMode
	}
	if svr.App.DevMode {
		t.extendParams["GetRules"] = func() []*Rule { return svr.rules }
		t.extendParams["GetStaticInfo"] = func() map[string]string {
			statics := map[string]string{}
			if svr.App.StaticInfo != nil {
				statics[svr.App.StaticInfo.UriPrefix] = svr.App.StaticInfo.Path
			}
			return statics
		}
	}
}

//reset template loader, clean template and loadedtemplate
func (t *TemplateLoader) reset() {
	t.template = tmpl.New("basename")
	t.loadedTemplate = map[string]string{}
}

func (t *TemplateLoader) load() error {
	for _, base := range t.basePath {

		template := t.template
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				ERROR.Println("Path Walk error:", err)
				return err
			}
			if info.IsDir() {
				if t.skipDir(info.Name()) {
					return filepath.SkipDir
				}
				return nil
			}

			if t.skipFile(info.Name()) {
				return nil
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			tmplName := templateName(path[len(base)+1:])
			//template already loaded
			if _, ok := t.loadedTemplate[tmplName]; ok {
				return nil
			}
			t.loadedTemplate[tmplName] = path
			_, err = template.New(tmplName).Funcs(t.FuncMap).Parse(string(data))
			if err != nil {
				ERROR.Println("template parse error:", err)
				complieError := &Error{
					SourceLines: strings.Split(string(data), "\n"),
					Title:       "Template compile error",
					FileName:    tmplName,
					Path:        path,
				}
				complieError.Line, complieError.Description = parseComplieError(err.Error())
				return complieError
			}
			return nil
		})
		if err != nil {
			if os.IsNotExist(err) {
				WARN.Println("template path is not exist:", base)
				continue
			}
			return err
		}

	}
	t.forceNotify = false
	return nil
}

var ErrorRE = regexp.MustCompile(`:(\d+):`)

//parse the template compile error
func parseComplieError(errstr string) (line int, desc string) {
	loc := ErrorRE.FindStringIndex(errstr)
	desc = errstr
	if loc != nil {
		var err error
		line, err = strconv.Atoi(errstr[loc[0]+1 : loc[1]-1])
		if err == nil {
			desc = errstr[loc[1]:]
		}
	}
	return
}

func (t *TemplateLoader) Lookup(path string) *tmpl.Template {
	if t.template == nil {
		return nil
	}
	template := t.template.Lookup(templateName(path))
	return template
}
