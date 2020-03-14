package struct2ts

import (
	"bytes"
	"testing"
)

func TestTabScanner_Write(t *testing.T) {
	tests := []struct {
		Name  string
		Input struct {
			Tabs    string
			Buffers []string
		}
		Expected struct {
			Output string
			n      *int
		}
	}{
		{
			"One new line mid string - one input, 3 spaces",
			struct {
				Tabs    string
				Buffers []string
			}{Tabs: "   ", Buffers: []string{"Hello how are you?\nI'm fine thanks!"}},
			struct {
				Output string
				n      *int
			}{Output: "Hello how are you?\n   I'm fine thanks!"},
		},
		{
			"One new line end string - one input, 3 spaces",
			struct {
				Tabs    string
				Buffers []string
			}{Tabs: "   ", Buffers: []string{"Hello how are you?\n"}},
			struct {
				Output string
				n      *int
			}{Output: "Hello how are you?\n   "},
		},
		{
			"One new line start string - one input, 3 spaces",
			struct {
				Tabs    string
				Buffers []string
			}{Tabs: "   ", Buffers: []string{"\nHello how are you?"}},
			struct {
				Output string
				n      *int
			}{Output: "\n   Hello how are you?"},
		},
		{
			"empty string",
			struct {
				Tabs    string
				Buffers []string
			}{Tabs: "   ", Buffers: []string{""}},
			struct {
				Output string
				n      *int
			}{Output: ""},
		},
		{
			"Just a new line",
			struct {
				Tabs    string
				Buffers []string
			}{Tabs: "   ", Buffers: []string{"\n"}},
			struct {
				Output string
				n      *int
			}{Output: "\n   "},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			outputBuffer := bytes.NewBuffer(nil)
			ts := newTabScanner(outputBuffer, test.Input.Tabs)
			for _, s := range test.Input.Buffers {
				_, err := ts.Write([]byte(s))
				if err != nil {
					t.Error("Buffer write error", err)
				}
			}
			if outputBuffer.String() != test.Expected.Output {
				t.Log("Output buffer doesn't match. Got ")
				t.Log(outputBuffer.String())
				t.Log("expected")
				t.Log(test.Expected.Output)
				t.Fail()
			}
		})
	}
}
