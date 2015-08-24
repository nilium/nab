package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	htx "html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	ttx "text/template"
)

type Template interface {
	Execute(io.Writer, interface{}) error
}

var funcs = map[string]interface{}{
	"nl": func() string { return "\n" },
}

var newline = []byte{'\n'}

func main() {
	useHTML := false
	flag.BoolVar(&useHTML, "html", false, "whether to use html/template instead of text/template")

	flag.Parse()

	log.SetPrefix("nab: ")
	log.SetFlags(0)

	var templates []Template
	for i, tx := range flag.Args() {
		var t Template
		name := fmt.Sprintf("tx:%d", i+1)
		if useHTML {
			t = htx.Must(htx.New(name).Funcs(funcs).Parse(tx))
		} else {
			t = ttx.Must(ttx.New(name).Funcs(funcs).Parse(tx))
		}

		templates = append(templates, t)
	}

	buf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	buf = bytes.Trim(buf, "\r\n \t")
	r := json.NewDecoder(bytes.NewBuffer(buf))

	var data interface{}
	if len(buf) >= 0 {
		switch buf[0] {
		case '"':
			var s string
			err = r.Decode(&s)
			data = s
		case 'n': // nil?
		case 't', 'f':
			var b bool
			err = r.Decode(&b)
			data = b
		case '{': // obj
			var m map[string]interface{}
			err = r.Decode(&m)
			data = m
		case '[': // array
			var slice []interface{}
			err = r.Decode(&slice)
			data = slice
		default: // number
			var f float64
			err = r.Decode(&f)
			data = f
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	last := len(templates) - 1
	for nth, t := range templates {
		if err := t.Execute(os.Stdout, data); err != nil {
			log.Fatal(err)
		}
		if last != nth {
			if _, err := os.Stdout.Write(newline); err != nil {
				log.Fatal(err)
			}
		}
	}
}
