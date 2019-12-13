// +build ignore

package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/zippopotamus/zippopotamus/internal"
	"os"
	"strings"
	"text/template"
	"time"
)

type staticFileType struct {
	name   string
	target string
	key    uint
	value  uint
}

var files = []staticFileType{
	{
		name:   "countries.txt",
		target: "Countries",
		key:    0,
		value:  4,
	},
	{
		name:   "admin_1.txt",
		target: "Admin1",
		key:    0,
		value:  1,
	},
	{
		name:   "admin_2.txt",
		target: "Admin2",
		key:    0,
		value:  1,
	},
}

func main() {
	log := logrus.StandardLogger()

	for _, f := range files {
		file, err := os.Open(f.name)

		if err != nil {
			log.WithError(err).Errorf("Unable to open file %s, continuing...", f.name)
			continue
		}

		defer file.Close()

		parser := internal.NewTsvParser(&internal.TsvParserOptions{
			Routines: 1,
			Logger:   log,
		})

		countryMap := make(internal.CodeMap)

		parser.Run(file, func(strings []string) {

			id := strings[f.key]
			value := strings[f.value]

			countryMap[id] = value
		})

		filename := fmt.Sprintf("./internal/static_%s.go", strings.ToLower(f.target))
		log.Infof("Writing to %s", filename)
		outputf, err := os.Create(filename)
		defer outputf.Close()

		outputTemplate.Execute(outputf, struct {
			Timestamp time.Time
			Target    string
			Countries internal.CodeMap
		}{
			Timestamp: time.Now(),
			Target:    f.target,
			Countries: countryMap,
		})

	}
}

var outputTemplate = template.Must(template.New("").Parse(`// Code generated; DO NOT EDIT.
// This file was generated at
// {{ .Timestamp }}
package internal

var {{.Target}} = CodeMap{
{{range $index, $element := .Countries }}
	{{printf "%q" $index}}: {{ printf "%q" $element }},{{end}}
}
`))
