package main

import (
	"fmt"
	"testing"
)

func TestCliFileSpec(t *testing.T) {
	a, b := "/a/b/c", FileSpec{"/a/b/c", "file", "", ""}
	if res, err := parseFileSpec(a); err != nil || res != b {
		t.Fatalf("%s != %s", b, res)
	}

	a, b = "alias=1,/a/b/c", FileSpec{"/a/b/c", "file", "1", ""}
	if res, err := parseFileSpec(a); err != nil || res != b {
		t.Fatalf("%s != %s", b, res)
	}

	a, b = "alias=2,/var/log/*.log", FileSpec{"/var/log/*.log", "glob", "2", ""}
	if res, err := parseFileSpec(a); err != nil || res != b {
		t.Fatalf("%s != %s", b, res)
	}

	a, b = "alias=1,group=\"a b\",/var/log/", FileSpec{"/var/log/", "dir", "1", "a b"}
	if res, err := parseFileSpec(a); err != nil || res != b {
		t.Fatalf("%s != %s", b, res)
	}
}

func getAliases(entries []*ListEntry) []string {
	aliases := make([]string, len(entries))
	for n, entry := range entries {
		aliases[n] = entry.Alias
	}
	return aliases
}

func TestListingWildcard(t *testing.T) {
	spec, _ := parseFileSpec("testdata/ex1/var/log/*.log")
	lst := createListing([]FileSpec{spec})

	if len(lst["__default__"]) != 4 {
		t.Fatalf("len(%#v) != 4\n", lst)
	}

	aliases := getAliases(lst["__default__"])
	if fmt.Sprintf("%q", aliases) != `["" "" "" ""]` {
		t.Fatal()
	}

	spec, _ = parseFileSpec("alias=logs,testdata/ex1/var/log/*.log")
	lst = createListing([]FileSpec{spec})

	aliases = getAliases(lst["__default__"])
	expect := `["logs/1.log" "logs/2.log" "logs/3.log" "logs/4.log"]`
	if fmt.Sprintf("%q", aliases) != expect {
		t.Fatalf("%q != %q", aliases, expect)
	}

}

func TestListingFile(t *testing.T) {
	spec1, _ := parseFileSpec("testdata/ex1/var/log/1.log")
	spec2, _ := parseFileSpec("testdata/ex1/var/log/2.log")
	lst := createListing([]FileSpec{spec1, spec2})

	aliases := getAliases(lst["__default__"])
	repr := fmt.Sprintf("%#q", aliases)
	if repr != "[`testdata/ex1/var/log/1.log` `testdata/ex1/var/log/2.log`]" {
		t.Fatal()
	}

	spec1, _ = parseFileSpec("group=a,alias=a.log,testdata/ex1/var/log/1.log")
	spec2, _ = parseFileSpec("group=b,alias=b.log,testdata/ex1/var/log/2.log")

	lst = createListing([]FileSpec{spec1, spec2})
	if lst["a"][0].Alias != "a.log" || !lst["a"][0].Exists {
		t.Fatal()
	}
	if lst["b"][0].Alias != "b.log" || !lst["b"][0].Exists {
		t.Fatal()
	}
}
