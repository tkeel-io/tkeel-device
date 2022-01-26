package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
)

type Template struct {
	mode string
	tmpl *template.Template
}

func New(templatePath, mode string) *Template {
	SwagdownTemplates := filepath.Join(filepath.Dir(templatePath), "*")
	funcMap := template.FuncMap{
		"FilterParameters": FilterParameters,
		"FilterSchema":     FilterSchema,
		"CollectSchema":    CollectSchema,
		"FormatAnchor":     FormatAnchor,
	}
	for name, fn := range sprig.FuncMap() {
		funcMap[name] = fn
	}

	if templatePath == "" {
		return &Template{
			mode,
			template.Must(template.New(fmt.Sprintf("%s.md", mode)).Funcs(funcMap).ParseFS(f, "templates/*")),
		}
	} else {
		return &Template{
			mode,
			template.Must(template.New(fmt.Sprintf("%s.md", mode)).Funcs(funcMap).ParseGlob(SwagdownTemplates)),
		}
	}
}

func RenderFromJSON(w Writer, r io.Reader, tmpl *Template) error {
	return renderAPI(tmpl, w, r, DecodeJSON)
}

func RenderFromYAML(w Writer, r io.Reader, tmpl *Template) error {
	return renderAPI(tmpl, w, r, DecodeYAML)
}

func renderAPI(tmpl *Template, w Writer, r io.Reader, decode func(data []byte) (*API, error)) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	api, err := decode(buf.Bytes())
	if err != nil {
		return err
	}

	switch tmpl.mode {
	case "tag": // 文档目录
		return renderTag(tmpl, w, api)
	case "method": // 文档函数
		return renderMethod(tmpl, w, api)
	}

	return errors.New("error mode")
}

func renderMethod(tmpl *Template, w Writer, api *API) error {
	tags := parseTags(api)
	for _, tag := range tags {
		//tagName := tag.Tag
		for _, method := range tag.Methods {
			filename := method.OperationID
			if filename == "" {
				continue
			}
			if err := tmpl.tmpl.Execute(w.For(fmt.Sprintf("%s_%s.md", "method", filename)), method); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderTag(tmpl *Template, w Writer, api *API) error {
	tags := parseTags(api)
	if err := tmpl.tmpl.Execute(w.For(fmt.Sprintf("tag.md")), tags); err != nil {
		return err
	}
	return nil
}

func parseTags(api *API) map[string]*Tag {
	tags := make(map[string]*Tag)
	for path, methods := range api.Paths {
		for typ, method := range methods {

			if len(method.Tags) == 0 {
				fmt.Printf("skip:method(%v) without tag\n", method.Summary)
				continue
			}

			key := method.Tags[0]
			if key == "" {
				fmt.Printf("error:method(%v) tag empty\n", method.Summary)
				continue
			}

			if method.OperationID == "" {
				fmt.Printf("error:tag(%v)method(%v) OperationID empty\n", method.Tags, method.Summary)
				continue
			}

			tag, ok := tags[key]
			if !ok {
				tag = &Tag{Tag: key, Methods: make([]*Operation, 0)}
				tags[key] = tag
			}
			method.Definitions = api.Definitions
			method.Operation = typ
			method.Path = filepath.Join(api.BasePath, path)
			tag.Methods = append(tag.Methods, method)
		}
	}
	return tags
}
