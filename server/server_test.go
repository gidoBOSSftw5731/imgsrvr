package server

//lol this is almost certainly broken

import (
	"path/filepath"
	"testing"
)

// TestReadKeys tests the readKeys function.
func TestReadKeys(t *testing.T) {
	// This is a typical 'table driven test' methodology for GoLang testing.
	// Define a 'table' of test data (slice of struct) and then for each test
	// run the required function (readKeys), and test for the proper result.
	tests := []struct {
		desc    string
		file    string
		testKey string
		want    bool
		wantErr bool
	}{{
		desc:    "Success - read single key", // Attempt to read a well formed file
		file:    "good_testkeys",             // and find a single key.
		testKey: "foo",
		want:    true,
	}, {
		desc:    "Fail - read single key", // Attempt to read a well formed file
		file:    "good_testkeys",          // and read a non-existent key.
		testKey: "blar",
		want:    false,
	}, {
		desc:    "Fail - read bad file", // Attempt to read a file which does not exist.
		file:    "nonfile",
		wantErr: true,
	}}

	// Now, loop over the tests and perform tests.
	for _, test := range tests {
		err := readKeys(filepath.Join("testdata", test.file))

		// This switch statement is a typical table-driven-test method as well.
		// Check that the test fails or not as expected, error if the incorrect
		// error state is achieved (err or no-error), then evaluate the final
		// test result: "got == want" ?? if that fails error appropriately.
		switch {
		case err != nil && !test.wantErr:
			t.Errorf("[%v] got an error when not expecting one: %v", test.desc, err)
		case err == nil && test.wantErr:
			t.Errorf("[%v] did not get an error when expecting one: %v", test.desc)
		case err == nil:
			// Arguably this test should just reflect.DeepEqual() the map of keys
			// to a defined set from the test instead of calling 'checkKey()' to validate
			// that the test data was loaded. That's a bit harder to do given the global nature of
			// the keys map. Testing checkKey() is just accessing the map element and should
			// be 'good enough' for this case though.
			got := checkKey(test.testKey)
			if got != test.want {
				t.Errorf("[%v] test failed got/want mismatch: %v/%v", test.desc, got, test.want)
			}
		}
	}
}

// TestCheckKey should test that a key is in the filled keys() map.
func TestCheckKey(t *testing.T) {
}
