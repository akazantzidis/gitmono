package gitmono

import (
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
)

// whitebox testing for autotag bump interface
func TestMinorBumper(t *testing.T) {
	for k, v := range map[string]string{
		"1":                  "1.1.0",
		"1.0":                "1.1.0",
		"1.0.0":              "1.1.0",
		"1.0.12":             "1.1.0",
		"1.0.0-patch":        "1.1.0",
		"1.0.0+build123":     "1.1.0",
		"1.0.0+build123.foo": "1.1.0",
		"1.0.0.0":            "1.1.0",
	} {
		tv, err := version.NewVersion(k)
		checkFatal(t, err)

		nv, err := minorBumper.bump(tv)
		checkFatal(t, err)

		if nv.String() != v {
			t.Fatalf("Expected '%s' got '%s'", v, nv.String())
		}
	}
}

func TestPatchBumper(t *testing.T) {
	// in retro this didn't have to be a map, but w/e
	for k, v := range map[string]string{
		"1":                      "1.0.1",
		"1.0":                    "1.0.1",
		"1.0.0":                  "1.0.1",
		"1.0.0-patch":            "1.0.1",
		"1.0.0+build123":         "1.0.1",
		"1.0.0+build123.foo.bar": "1.0.1",
	} {
		tv, err := version.NewVersion(k)
		checkFatal(t, err)

		nv, err := patchBumper.bump(tv)
		checkFatal(t, err)

		if nv.String() != v {
			t.Fatalf("Expected '%s' got '%s'", v, nv.String())
		}
	}
}

func TestMajorBumper(t *testing.T) {
	for k, v := range map[string]string{
		"1":                  "2.0.0",
		"1.0":                "2.0.0",
		"1.1":                "2.0.0",
		"1.0.0":              "2.0.0",
		"1.1.0":              "2.0.0",
		"1.0.0-patch":        "2.0.0",
		"1.0.0+build123":     "2.0.0",
		"1.0.0+build123.foo": "2.0.0",
		"1.0.12":             "2.0.0",
	} {
		tv, err := version.NewVersion(k)
		checkFatal(t, err)

		nv, err := majorBumper.bump(tv)
		checkFatal(t, err)

		if nv.String() != v {
			t.Fatalf("Expected '%s' got '%s'", v, nv.String())
		}
	}
}

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}
