package peony

import (
	"errors"
	tmpl "html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	NotDirError = errors.New("Is not a directory")
)

type TemplateLoader struct {
	template *tmpl.Template //
}

func (t *TemplateLoader) skipDir(n string) bool {
	return n[0] == '.'
}

func (t *TemplateLoader) skipFile(n string) bool {
	return n[0] == '.'
}

func templateName(path string) string {
	if os.PathSeparator == '\\' {
		return strings.Replace(path, string(os.PathSeparator), ".", -1)
	}
	return path
}

func NewTemplateLoader() (*TemplateLoader, error) {
	tl := &TemplateLoader{}
	return tl, nil
}

func (t *TemplateLoader) load(base string) error {
	basePath, err := filepath.Abs(base)
	if err != nil {
		return err
	}
	var bPathFileInfo os.FileInfo
	if bPathFileInfo, err = os.Stat(basePath); err != nil {
		return err
	}
	if !bPathFileInfo.IsDir() {
		return NotDirError
	}
	if err != nil {
		return err
	}
	if t.template == nil {
		t.template = tmpl.New("root")
	}
	basename := templateName(basePath)
	template := t.template.Lookup(basename)
	if template == nil {
		template = t.template.New(basename)
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
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
		tmplName := templateName(path[len(basePath)+1:])
		_, err = template.New(tmplName).Parse(string(data))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *TemplateLoader) Lookup(base string) *tmpl.Template {
	basePath, _ := filepath.Abs(base)
	template := t.template.Lookup(templateName(basePath))
	return template
}
