package main

import (
	"html/template"
)

// Шаблоны, используемые для преобразования информации в публикацию.
var templates = template.Must(template.New("").Parse(`
{{ define "header"}}<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="{{ if .lang }}{{ .lang }}{{ else }}en{{ end }}">
<head>
<meta charset="UTF-8" />
<title>{{ .title }}</title>{{ if ._globalcssfile_ }}
<link rel="stylesheet" href="{{ ._globalcssfile_ }}" />{{ end }}
</head>
<body{{ if .class }} class="{{ .class }}"{{ end }}>{{ end }}

{{ define "footer" }}</body>
</html>{{ end }}

{{ define "page" }}{{ template "header" . }}
{{ .content }}{{ template "footer" }}{{ end }}

{{ define "toc" }}{{ template "header" . }}
<nav epub:type="toc">
<ol>{{ range .toc }}
<li><a href="{{ .Filename }}">{{ if .Title }}{{ .Title }}{{ else }}* * *{{ end }}</a></li>{{ end }}
</ol>
</nav>
{{ template "footer" }}{{ end }}

{{ define "nav" }}{{ template "header" . }}
<nav epub:type="toc">
{{ .content }}
</nav>
{{ template "footer" }}{{ end }}`))
