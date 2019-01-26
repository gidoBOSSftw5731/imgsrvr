package server

import (
	"path/filepath"
	"testing"
)

func TestReadKeys(t *testing.T) {
	tests := []struct {
		desc    string
		file    string
		testKey string
		want    bool
		wantErr bool
	}{{
		desc:    "Success - read single key",
		file:    "good_testkeys",
		testKey: "foo",
		want:    true,
	}, {
		desc:    "Fail - read single key",
		file:    "good_testkeys",
		testKey: "blar",
		want:    false,
	}, {
		desc:    "Fail - read bad file",
		file:    "nonfile",
		wantErr: true,
	}}

	for _, test := range tests {
		err := readKeys(filepath.Join("testdata", test.file))
		switch {
		case err != nil && !test.wantErr:
			t.Errorf("[%v] got an error when not expecting one: %v", test.desc, err)
		case err == nil && test.wantErr:
			t.Errorf("[%v] did not get an error when expecting one: %v", test.desc)
		case err == nil:
			got := checkKey(test.testKey)
			if got != test.want {
				t.Errorf("[%v] test failed got/want mismatch: %v/%v", test.desc, got, test.want)
			}
		}
	}
}
func TestCheckKey(t *testing.T) {
}
