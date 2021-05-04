# go-giter8
[![Go](https://github.com/Gympass/go-giter8/actions/workflows/go.yml/badge.svg)](https://github.com/Gympass/go-giter8/actions/workflows/go.yml)

`go-giter8` implements a simple library capable of handling [giter8](https://github.com/foundweekends/giter8) templates.

## Using as library

1. Add to `go.mod`
```
module github.com/yourusername/yourproject

go 1.16

require github.com/Gympass/go-giter8 v0.1.1
```

2. Use `props` to parse `default.properties`:

```go
package foo

import (
	"io"
	"os"

	"github.com/Gympass/go-giter8/props"
)

func parseProperties() props.Pairs {
	f, _ := os.Open("/path/to/your/default.properties")
	data, _ := io.ReadAll(f)
	properties, err := props.ParseProperties(string(data))
	if err != nil {
		panic(err)
    }
    return properties
}
```

3. Use parsed properties in a template

```go
package foo

import (
	"io"
	"os"

	"github.com/Gympass/go-giter8/lexer"
	"github.com/Gympass/go-giter8/props"
	"github.com/Gympass/go-giter8/render"
)

func parseProperties() props.Pairs { /* ... */ }

func executeTemplate(path string) (string, error) {
	f, _ := os.Open("/path/to/template/file")
	data, _ := io.ReadAll(f)
	parsed, err := lexer.Tokenize(string(data))
	if err != nil {
		return "", err
	}
	e := render.NewExecutor(parseProperties())
	return e.Exec(parsed)
}
```

## Using as command line
Alternatively, you can use the `gg8` CLI to download and execute a template:

```bash
$ gg8 Gympass/test.g8 test
Clonning https://github.com/Gympass/test.g8.git...
Executing template:

name [Test project]: <return>
otherValue [Pizza]: <return>
Processing templates... OK

$   
```

One can also provide all options through command-line parameters:

```bash
$ gg8 Gympass/test.g8 test -- name="A name" otherValue="othervalue"
Clonning https://github.com/Gympass/test.g8.git...
Executing template:
name: A name 
otherValue: othervalue

Processing templates... OK
```

## License

```
MIT License

Copyright (c) 2021 Victor Gama

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
