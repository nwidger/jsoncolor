package jsoncolor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

func Marshal(v interface{}) ([]byte, error) {
	f := NewFormatter()
	return marshal(v, f)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	f := NewFormatter()
	f.Prefix = prefix
	f.Indent = indent
	return marshal(v, f)
}

func marshal(v interface{}, f *Formatter) ([]byte, error) {
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
	DefaultObjectColor = color.New(color.FgWhite, color.Bold)
	DefaultArrayColor  = color.New(color.FgWhite, color.Bold)
	DefaultFieldColor  = color.New(color.FgBlue, color.Bold)
	DefaultStringColor = color.New(color.FgGreen)
	DefaultTrueColor   = color.New(color.FgWhite)
	DefaultFalseColor  = color.New(color.FgWhite)
	DefaultNumberColor = color.New(color.FgWhite)
	DefaultNullColor   = color.New(color.FgBlack, color.Bold)

	DefaultPrefix = ""
	DefaultIndent = "  "
)

type Formatter struct {
	ObjectColor *color.Color
	ArrayColor  *color.Color
	FieldColor  *color.Color
	StringColor *color.Color
	TrueColor   *color.Color
	FalseColor  *color.Color
	NumberColor *color.Color
	NullColor   *color.Color

	Prefix string
	Indent string
}

func NewFormatter() *Formatter {
	return &Formatter{
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

func (f *Formatter) Format(dst *bytes.Buffer, src []byte) error {
	return newFormatterState(f, dst).format(dst, src)
}

type formatterState struct {
	indent string
	frames []*frame

	printDelim  func(json.Delim)
	printField  func(k string) error
	printString func(s string) error
	printBool   func(b bool)
	printNumber func(n json.Number)
	printNull   func()
	printIndent func()
}

func newFormatterState(f *Formatter, dst *bytes.Buffer) *formatterState {
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
		printDelim: func(t json.Delim) {
			if t == json.Delim('{') || t == json.Delim('}') {
				fmt.Fprint(dst, sprintfObject("%v", t))
			} else {
				fmt.Fprint(dst, sprintfArray("%v", t))
			}
		},
		printField: func(k string) error {
			sbuf, err := json.Marshal(&k)
			if err != nil {
				return err
			}
			fmt.Fprint(dst, sprintfField("%v", string(sbuf)))
			return nil
		},
		printString: func(s string) error {
			sbuf, err := json.Marshal(&s)
			if err != nil {
				return err
			}
			fmt.Fprint(dst, sprintfString("%v", string(sbuf)))
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
			fmt.Fprint(dst, sprintfNull("%v", "null"))
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
			fmt.Fprint(dst, fs.indent[:ilen])
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
		fs.printDelim(x)
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

		end := "\n"
		if frame.inArrayOrObject() && dec.More() {
			end = ",\n"
		}

		if x, ok := t.(json.Delim); ok {
			if x == json.Delim('{') || x == json.Delim('[') {
				if frame.inObject() {
					fmt.Fprint(dst, " ")
				} else {
					fs.printIndent()
				}
				err = fs.formatToken(x)
				more := dec.More()
				if more {
					fmt.Fprint(dst, "\n")
				}
				frame = fs.enterFrame(x, !dec.More())
			} else {
				empty := frame.isEmpty()
				frame = fs.leaveFrame()
				if !empty {
					fs.printIndent()
				}
				err = fs.formatToken(x)
				fmt.Fprint(dst, end)
			}
		} else {
			printIndent := frame.inArray()
			if _, ok := t.(string); ok {
				printIndent = !frame.inObject() || frame.inField()
				if frame.inField() {
					end = ": "
				}
			}

			if printIndent {
				fs.printIndent()
			}
			err = fs.formatToken(t)
			fmt.Fprint(dst, end)
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
