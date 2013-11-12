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
	template     *tmpl.Template //
	basePath     string
	compileError error
	forceNotify  bool //force notify when the first time
}

func (t *TemplateLoader) IgnoreDir(file os.FileInfo) bool {
	if strings.HasPrefix(file.Name(), ".") {
		return true
	}
	return false
}

//when the suffix is .html notify the change
func (t *TemplateLoader) IgnoreFile(file string) bool {
	if !strings.HasSuffix(filepath.Base(file), ".go") {
		return true
	}
	return false
}

func (t *TemplateLoader) Refresh() error {
	t.template = nil
	t.compileError = nil
	return t.load()
}

func (t *TemplateLoader) ForceRefresh() bool {
	if t.forceNotify {
		t.forceNotify = false
		return true
	}
	return t.forceNotify
}

func (t *TemplateLoader) Path() string {
	return t.basePath
}

func (t *TemplateLoader) skipDir(n string) bool {
	return strings.HasPrefix(n, ".")
}

func (t *TemplateLoader) skipFile(n string) bool {
	return strings.HasPrefix(n, ".")
}

func templateName(path string) string {
	if os.PathSeparator == '\\' {
		return strings.Replace(path, string(os.PathSeparator), ".", -1)
	}
	return path
}

func NewTemplateLoader(base string) *TemplateLoader {
	tl := &TemplateLoader{basePath: base, forceNotify: true}
	return tl
}

func (t *TemplateLoader) load() error {
	basename := templateName(t.basePath)
	if t.template == nil {
		t.template = tmpl.New(basename)
	}
	template := t.template
	err := filepath.Walk(t.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
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
		tmplName := templateName(path[len(t.basePath)+1:])
		_, err = template.New(tmplName).Parse(string(data))
		if err != nil {
			complieError := &Error{
				SouceLines: strings.Split(string(data), "\n"),
				Title:      "Template compile error",
				FileName:   tmplName,
				Path:       path,
			}
			complieError.Line, complieError.Description = parseComplieError(err.Error())
			t.compileError = complieError
			ERROR.Println(err)
			return t.compileError
		}
		return nil
	})
	if err != nil {
		return err
	}
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

func (t *TemplateLoader) Lookup(path string) (*tmpl.Template, error) {
	if t.compileError != nil {
		return nil, t.compileError
	}
	template := t.template.Lookup(templateName(path))
	var err = NoTemplateFound
	if template != nil {
		err = nil
	}
	return template, err
}
