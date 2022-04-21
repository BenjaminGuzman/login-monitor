package main

import (
	"reflect"
	"testing"
)

func TestWrap(t *testing.T) {
	type args struct {
		src    []byte
		maxLen int
		sep    string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"Test no wrap", args{[]byte{1, 2, 3}, 5, "ab"}, []byte{1, 2, 3}},
		{"Test single wrap", args{[]byte{1, 2, 3}, 2, "a"}, []byte{1, 2, 'a', 3}},
		{"Test multiple wrap", args{[]byte{1, 2, 3, 4, 5, 6, 7}, 3, "a"}, []byte{1, 2, 3, 'a', 4, 5, 6, 'a', 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Wrap(tt.args.src, tt.args.maxLen, tt.args.sep); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Wrap() = %v, want %v", got, tt.want)
			}
		})
	}
}
