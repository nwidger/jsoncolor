// Package jsoncolor is a replacement for encoding/json's
// MarshalIndent producing colorized output.
package jsoncolor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

// MarshalIndent is like encoding/json's MarshalIndent but colorizes
// the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	f := NewFormatter()
	f.Prefix = prefix
	f.Indent = indent

	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(b)))
	err = f.Format(buf, b)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type frame struct {
	object bool
	field  bool
	array  bool
	empty  bool
	indent int
}

func (f *frame) inArray() bool {
	if f == nil {
		return false
	}
	return f.array
}

func (f *frame) inObject() bool {
	if f == nil {
		return false
	}
	return f.object
}

func (f *frame) inArrayOrObject() bool {
	if f == nil {
		return false
	}
	return f.object || f.array
}

func (f *frame) inField() bool {
	if f == nil {
		return false
	}
	return f.object && f.field
}

func (f *frame) toggleField() {
	if f == nil {
		return
	}
	f.field = !f.field
}

func (f *frame) isEmpty() bool {
	if f == nil {
		return false
	}
	return (f.object || f.array) && f.empty
}

var (
	DefaultSpaceColor  = color.New()
	DefaultCommaColor  = color.New(color.FgWhite)
	DefaultColonColor  = color.New(color.FgWhite)
	DefaultObjectColor = color.New(color.FgWhite, color.Bold)
	DefaultArrayColor  = color.New(color.FgWhite, color.Bold)
	DefaultFieldColor  = color.New(color.FgBlue, color.Bold)
	DefaultStringColor = color.New(color.FgGreen)
	DefaultTrueColor   = color.New(color.FgWhite)
	DefaultFalseColor  = color.New(color.FgWhite)
	DefaultNumberColor = color.New(color.FgWhite)
	DefaultNullColor   = color.New(color.FgBlack, color.Bold)

	// By default, no prefix is used.
	DefaultPrefix = ""
	// By default, an indentation of two spaces is used.
	DefaultIndent = "  "
)

func init() {
	DefaultSpaceColor.DisableColor()
}

// Formatter colorizes buffers containing JSON.
type Formatter struct {
	// Color for whitespace characters.  DisableColor is called on
	// DefaultSpaceColor in a package init function so that
	// whitespace characters are not colored by default.
	SpaceColor *color.Color
	// Color for comma character ',' delimiting object and array
	// fields.
	CommaColor *color.Color
	// Color for colon character ':' separating object field names
	// and values.
	ColonColor *color.Color
	// Color for object delimiter characters '{' and '}'.
	ObjectColor *color.Color
	// Color for array delimiter characters '[' and ']'.
	ArrayColor *color.Color
	// Color for object field names.
	FieldColor *color.Color
	// Color for string values.
	StringColor *color.Color
	// Color for 'true' boolean values.
	TrueColor *color.Color
	// Color for 'false' boolean values.
	FalseColor *color.Color
	// Color for number values.
	NumberColor *color.Color
	// Color for null values.
	NullColor *color.Color

	// Prefix is prepended before indentation to newlines.
	Prefix string
	// Indent is prepended to newlines one or more times according
	// to indentation nesting.
	Indent string
}

// NewFormatter returns a new formatter.
func NewFormatter() *Formatter {
	return &Formatter{
		SpaceColor:  DefaultSpaceColor,
		CommaColor:  DefaultCommaColor,
		ColonColor:  DefaultColonColor,
		ObjectColor: DefaultObjectColor,
		ArrayColor:  DefaultArrayColor,
		FieldColor:  DefaultFieldColor,
		StringColor: DefaultStringColor,
		TrueColor:   DefaultTrueColor,
		FalseColor:  DefaultFalseColor,
		NumberColor: DefaultNumberColor,
		NullColor:   DefaultNullColor,
		Prefix:      DefaultPrefix,
		Indent:      DefaultIndent,
	}
}

// Format appends to dst a colorized form of the JSON-encoded src.
func (f *Formatter) Format(dst *bytes.Buffer, src []byte) error {
	return newFormatterState(f, dst).format(dst, src)
}

type formatterState struct {
	indent string
	frames []*frame

	printSpace  func(string)
	printComma  func()
	printColon  func()
	printObject func(json.Delim)
	printArray  func(json.Delim)
	printField  func(k string) error
	printString func(s string) error
	printBool   func(b bool)
	printNumber func(n json.Number)
	printNull   func()
	printIndent func()
}

func newFormatterState(f *Formatter, dst *bytes.Buffer) *formatterState {
	sprintfSpace := f.SpaceColor.SprintfFunc()
	sprintfComma := f.CommaColor.SprintfFunc()
	sprintfColon := f.ColonColor.SprintfFunc()
	sprintfObject := f.ObjectColor.SprintfFunc()
	sprintfArray := f.ArrayColor.SprintfFunc()
	sprintfField := f.FieldColor.SprintfFunc()
	sprintfString := f.StringColor.SprintfFunc()
	sprintfTrue := f.TrueColor.SprintfFunc()
	sprintfFalse := f.FalseColor.SprintfFunc()
	sprintfNumber := f.NumberColor.SprintfFunc()
	sprintfNull := f.NullColor.SprintfFunc()

	fs := &formatterState{
		indent: "",
		frames: []*frame{
			{},
		},
		printSpace: func(s string) {
			fmt.Fprint(dst, sprintfSpace(s))
		},
		printComma: func() {
			fmt.Fprint(dst, sprintfComma(","))
		},
		printColon: func() {
			fmt.Fprint(dst, sprintfColon(":"))
		},
		printObject: func(t json.Delim) {
			fmt.Fprint(dst, sprintfObject(t.String()))
		},
		printArray: func(t json.Delim) {
			fmt.Fprint(dst, sprintfArray(t.String()))
		},
		printField: func(k string) error {
			sbuf, err := json.Marshal(&k)
			if err != nil {
				return err
			}
			fmt.Fprint(dst, sprintfField(string(sbuf)))
			return nil
		},
		printString: func(s string) error {
			sbuf, err := json.Marshal(&s)
			if err != nil {
				return err
			}
			fmt.Fprint(dst, sprintfString(string(sbuf)))
			return nil
		},
		printBool: func(b bool) {
			if b {
				fmt.Fprint(dst, sprintfTrue("%v", b))
			} else {
				fmt.Fprint(dst, sprintfFalse("%v", b))
			}
		},
		printNumber: func(n json.Number) {
			fmt.Fprint(dst, sprintfNumber("%v", n))
		},
		printNull: func() {
			fmt.Fprint(dst, sprintfNull("null"))
		},
	}

	fs.printIndent = func() {
		if len(f.Prefix) > 0 {
			fmt.Fprint(dst, f.Prefix)
		}
		indent := fs.frame().indent
		if indent > 0 {
			ilen := len(f.Indent) * indent
			if len(fs.indent) < ilen {
				fs.indent = strings.Repeat(f.Indent, indent)
			}
			fmt.Fprint(dst, sprintfSpace(fs.indent[:ilen]))
		}
	}

	return fs
}

func (fs *formatterState) frame() *frame {
	return fs.frames[len(fs.frames)-1]
}

func (fs *formatterState) enterFrame(t json.Delim, empty bool) *frame {
	indent := fs.frames[len(fs.frames)-1].indent + 1
	fs.frames = append(fs.frames, &frame{
		object: t == json.Delim('{'),
		array:  t == json.Delim('['),
		indent: indent,
		empty:  empty,
	})
	return fs.frame()
}

func (fs *formatterState) leaveFrame() *frame {
	fs.frames = fs.frames[:len(fs.frames)-1]
	return fs.frame()
}

func (fs *formatterState) formatToken(t json.Token) error {
	switch x := t.(type) {
	case json.Delim:
		if x == json.Delim('{') || x == json.Delim('}') {
			fs.printObject(x)
		} else {
			fs.printArray(x)
		}
	case json.Number:
		fs.printNumber(x)
	case string:
		if !fs.frame().inField() {
			return fs.printString(x)
		}
		return fs.printField(x)
	case bool:
		fs.printBool(x)
	case nil:
		fs.printNull()
	default:
		return fmt.Errorf("unknown type %T", t)
	}
	return nil
}

func (fs *formatterState) format(dst *bytes.Buffer, src []byte) error {
	dec := json.NewDecoder(bytes.NewReader(src))
	dec.UseNumber()

	frame := fs.frame()

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		more := dec.More()
		printComma := frame.inArrayOrObject() && more

		if x, ok := t.(json.Delim); ok {
			if x == json.Delim('{') || x == json.Delim('[') {
				if frame.inObject() {
					fs.printSpace(" ")
				} else {
					fs.printIndent()
				}
				err = fs.formatToken(x)
				if more {
					fs.printSpace("\n")
				}
				frame = fs.enterFrame(x, !more)
			} else {
				empty := frame.isEmpty()
				frame = fs.leaveFrame()
				if !empty {
					fs.printIndent()
				}
				err = fs.formatToken(x)
				if printComma {
					fs.printComma()
				}
				fs.printSpace("\n")
			}
		} else {
			printIndent := frame.inArray()
			if _, ok := t.(string); ok {
				printIndent = !frame.inObject() || frame.inField()
			}

			if printIndent {
				fs.printIndent()
			}
			if !frame.inField() {
				fs.printSpace(" ")
			}
			err = fs.formatToken(t)
			if frame.inField() {
				fs.printColon()
			} else {
				if printComma {
					fs.printComma()
				}
				fs.printSpace("\n")
			}
		}

		if frame.inObject() {
			frame.toggleField()
		}

		if err != nil {
			return err
		}
	}

	return nil
}
