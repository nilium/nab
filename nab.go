package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	},
	"prettyjson": func(v interface{}) (string, error) {
		b, err := json.MarshalIndent(v, "", "\t")
		return string(b), err
	},
	"unjson": func(d interface{}) (interface{}, error) {
		var b []byte
		switch d := d.(type) {
		case string:
			b = []byte(d)
		case []byte:
			b = d
		default:
			return nil, fmt.Errorf("Expected []byte or string, got %T", d)
		}

		if len(b) == 0 {
			return nil, errors.New("JSON data empty")
		}

		switch b[0] {
		case '{':
			r := map[string]interface{}{}
			return r, json.Unmarshal(b, &r)
		case '[':
			var r []interface{}
			return r, json.Unmarshal(b, &r)
		case '"':
			var r string
			return r, json.Unmarshal(b, &r)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+':
			var r float64
			return r, json.Unmarshal(b, &r)
		case 't', 'f':
			var r bool
			return r, json.Unmarshal(b, &r)
		case 'n':
			var r interface{}
			if string(b) != "null" {
				break
			}
			return nil, json.Unmarshal(b, &r)
		}

		return nil, errors.New("Unable to determine JSON root type")
	},
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
