package jsoncolor

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type object struct {
	Null   *string `json:"null"`
	True   bool    `json:"true"`
	False  bool    `json:"false"`
	Number int     `json:"number"`
	String string  `json:"string"`
	Array  []int   `json:"array"`
	Object *object `json:"object"`
}

func TestCompareWithStd(t *testing.T) {
	obj := object{
		True:   true,
		False:  false,
		Number: 123,
		String: "string",
		Array:  []int{1, 2, 3},
		Object: &object{
			True:   true,
			False:  false,
			Number: 123,
			String: "string",
			Array:  []int{1, 2, 3},
		},
	}

	tests := []struct {
		name           string
		v              interface{}
		prefix, indent string
	}{
		{name: "null", v: nil, prefix: "<p>", indent: "<i>"},
		{name: "true", v: true, prefix: "<p>", indent: "<i>"},
		{name: "false", v: false, prefix: "<p>", indent: "<i>"},
		{name: "number", v: 123, prefix: "<p>", indent: "<i>"},
		{name: "string", v: "string", prefix: "<p>", indent: "<i>"},
		{name: "array", v: []int{1, 2, 3}, prefix: "<p>", indent: "<i>"},
		{name: "object", v: obj, prefix: "<p>", indent: "<i>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := json.MarshalIndent(tt.v, tt.prefix, tt.indent)
			if err != nil {
				t.Fatal(err)
			}
			want := string(buf)
			buf2, err := MarshalIndent(tt.v, tt.prefix, tt.indent)
			if err != nil {
				t.Fatal(err)
			}
			got := string(buf2)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("MarshalIndent() mismatch (-want +got):\n%s", diff)
			}
			buf, err = json.Marshal(tt.v)
			if err != nil {
				t.Fatal(err)
			}
			want = string(buf)
			buf2, err = Marshal(tt.v)
			if err != nil {
				t.Fatal(err)
			}
			got = string(buf2)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
