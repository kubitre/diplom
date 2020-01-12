package utils

import (
	"bytes"
	"text/template"
)

const (
	templateExecutingScript = `
#!/bin/bash
# auto generated. Don't edit
{{range $v := .}}
{{$v}}
{{end}}
	`
)

func CreateExecutingScript(shellCommands []string) ([]byte, error) {
	templ := template.New("")
	templ, err := templ.Parse(templateExecutingScript)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err := templ.Execute(buf, shellCommands); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
