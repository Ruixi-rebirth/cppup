package template

import (
	"bytes"
	"embed"
	texttemplate "text/template"
)

//go:embed files
var templateFiles embed.FS

func Render(path string, data any) (string, error) {
	content, err := templateFiles.ReadFile(path)
	if err != nil {
		return "", err
	}
	t, err := texttemplate.New(path).Parse(string(content))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderStatic(path string) (string, error) {
	content, err := templateFiles.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
