jsoncolor
=========

[![GoDoc](https://godoc.org/github.com/nwidger/jsoncolor?status.svg)](https://godoc.org/github.com/nwidger/jsoncolor)

`jsoncolor` is a drop-in replacement for `encoding/json`'s
`MarshalIndent` function which produces colorized output using fatih's
[color](https://github.com/fatih/color) package.

## Installation

```
go get -u github.com/nwidger/jsoncolor
```

## Usage

To use as a replacement for `encoding/json`, exchange

`import "encoding/json"` with `import json "github.com/nwidger/jsoncolor"`.

`json.MarshalIndent` will now produce colorized output.

## Custom Colors

The colors used for each type of token can be customized by creating a
custom `Formatter`, changing its `XXXColor` fields and calling its
`Format` method.  See
[color.New](https://godoc.org/github.com/fatih/color#New) for creating
custom color types and
[jsoncolor.NewFormatter](https://godoc.org/github.com/nwidger/jsoncolor#NewFormatter)
for the default colors.

``` go
import (
        "bytes"
		"encoding/json"
		"fmt"
		"log"

        "github.com/fatih/color"
        "github.com/nwidger/jsoncolor"
)

// marshal v using stdlib
src, err := json.Marshal(v)
if err != nil {
        log.Fatal(err)
}

// create custom formatter,
// set custom colors
f := jsoncolor.NewFormatter()
f.StringColor = color.New(color.FgBlue, color.Bold)
f.NumberColor = color.New(color.FgWhite)

// colorized output is written to dst
dst := &bytes.Buffer{}
err := f.Format(dst, src)
if err != nil {
        log.Fatal(err)
}

fmt.Println(dst.String())
```
