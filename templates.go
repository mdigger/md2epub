package main

import (
	"html/template"
)

var templates = template.Must(template.New("").Parse(`
{{ define "header"}}<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="{{ if .lang }}{{ .lang }}{{ else }}en{{ end }}">
<head>
<meta charset="UTF-8" />
<title>{{ .title }}</title>
</head>
<body>{{ end }}

{{ define "footer" }}</body>
</html>{{ end }}

{{ define "page" }}{{ template "header" . }}
{{ .content }}
{{ template "footer" }}{{ end }}

{{ define "toc" }}{{ template "header" . }}
<nav epub:type="toc">
<h1>{{ .title }}</h1>
<ol>{{ range .toc }}
<li><a href="{{ .Filename }}">{{ if .Title }}{{ .Title }}{{ else }}* * *{{ end }}</a></li>{{ end }}
</ol>
</nav>
{{ template "footer" }}{{ end }}`))
