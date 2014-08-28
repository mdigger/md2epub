package main

import (
	"html/template"
)

var tnav = template.Must(template.New("").Parse(navTemplateText))

const navTemplateText = `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="{{ if .lang }}{{ .lang }}{{ else }}en{{ end }}">
<head>
<meta charset="UTF-8" />
<title>{{ .title }}</title>
</head>
<body>
<nav epub:type="toc">
<h1>{{ .title }}</h1>
<ol>{{ range .toc }}
<li><a href="{{ .Filename }}">{{ if .Title }}{{ .Title }}{{ else }}* * *{{ end }}</a></li>{{ end }}
</ol>
</nav>
</body>
</html>`
