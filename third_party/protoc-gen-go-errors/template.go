// MIT License
//
// Copyright (c) 2020 go-kratos
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"
)

const errorsTpl = `
{{ range .Errors }}
var {{.LowerCamelValue}} *errors.TError
{{- end }}

func init() {
{{- range .Errors }}
{{.LowerCamelValue}} = errors.New(int(codes.{{.Code}}), "{{.Key}}", {{.Name}}_{{.Value}}.String())
errors.Register({{.LowerCamelValue}})
{{- end }}
}

{{ range .Errors }}
func {{.UpperCamelValue}}() errors.Error {
	 return {{.LowerCamelValue}}
}
{{ end }}
`

type errorInfo struct {
	Name            string
	Value           string
	Code            string
	UpperCamelValue string
	LowerCamelValue string
	Key             string
}

type errorWrapper struct {
	Errors []*errorInfo
}

func (e *errorWrapper) execute() string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("errors").Parse(errorsTpl)
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, e); err != nil {
		panic(err)
	}
	return string(GoFmt(buf.Bytes()))
}

// GoFmt 格式化代码
func GoFmt(buf []byte) []byte {
	formatted, err := format.Source(buf)
	if err != nil {
		panic(fmt.Errorf("%s\nOriginal code:\n%s", err.Error(), buf))
	}
	return formatted
}
