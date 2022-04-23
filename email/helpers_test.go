package email

import (
	"os"
	"reflect"
	"testing"
	"time"
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

func TestReplacePlaceholders(t *testing.T) {
	timeTemplate := "hola %t02 Jan 06t% mundo %t Monday, 02-Jan-06 t% bye"
	fileTemplate := "hola %f helpers_test.go f% mundo. %fhelpers_test.gof% bye"
	expectedTimeReplacement := "hola " + time.Now().Format("02 Jan 06") + " mundo " + time.Now().Format("Monday, 02-Jan-06") + " bye"

	expectedFileReplacement := "hola "
	f, _ := os.ReadFile("helpers_test.go")
	expectedFileReplacement += string(f) + " mundo. " + string(f) + " bye"

	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Testing replacement of time", args{timeTemplate}, expectedTimeReplacement},
		{"Testing file replacement", args{fileTemplate}, expectedFileReplacement},
		{"Testing both replacements", args{timeTemplate + " " + fileTemplate}, expectedTimeReplacement + " " + expectedFileReplacement},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplacePlaceholders(tt.args.str); got != tt.want {
				t.Errorf("ReplacePlaceholders() = %v, want %v", got, tt.want)
			}
		})
	}
}
