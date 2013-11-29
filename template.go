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
	template    *tmpl.Template //
	basePath    []string
	forceNotify bool //force notify when the first time
}

func (t *TemplateLoader) IgnoreDir(file os.FileInfo) bool {
	if strings.HasPrefix(file.Name(), ".") {
		return true
	}
	return false
}

//when the suffix is .html notify the change
func (t *TemplateLoader) IgnoreFile(file string) bool {
	if !strings.HasSuffix(filepath.Base(file), ".html") {
		return true
	}
	return false
}

func (t *TemplateLoader) Refresh() error {
	t.template = nil
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
	tl := &TemplateLoader{basePath: base, forceNotify: true}
	return tl
}

var (
	TemplateFuncs = map[string]interface{}{
		"IsDevMode": func() bool { return DevMode },
	}
)

func (t *TemplateLoader) load() error {

	for _, base := range t.basePath {
		if t.template == nil {
			t.template = tmpl.New("basename")
		}
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
			_, err = template.New(tmplName).Funcs(TemplateFuncs).Parse(string(data))
			if err != nil {
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
