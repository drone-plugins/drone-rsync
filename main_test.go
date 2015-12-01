package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/drone/drone-go/drone"
)

func TestBuildRsync(t *testing.T) {
	testdata := []struct {
		host string
		root string
		rs   Rsync
		exp  []string
	}{
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: "testdata/*.txt",
				Target: "/dev/null",
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"testdata/bar.txt",
				"testdata/foo.txt",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: "testdata/foo*",
				Target: "/dev/null",
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"testdata/foo",
				"testdata/foo.txt",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: "testdata/bar-*.x86_64.rpm",
				Target: "/dev/null",
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"testdata/bar-0.0.30+g4cdc188-1.x86_64.rpm",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: "notthere/*.txt",
				Target: "/dev/null",
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"notthere/*.txt",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:      "jqhacker",
				Source:    "testdata/",
				Target:    "/dev/null",
				Recursive: true,
				Delete:    true,
			},
			[]string{
				"rsync",
				"-az",
				"-r",
				"--del",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"testdata/",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: "testdata/foo.txt",
				Target: "/dev/null",
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"testdata/foo.txt",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: "testdata/bar.txt",
				Target: "/dev/null",
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"testdata/bar.txt",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: abs("testdata/bar.txt"),
				Target: "/dev/null",
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				abs("testdata/bar.txt"),
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:    "jqhacker",
				Target:  "/dev/null",
				Source:  "./",
				Include: ss([]string{"testdata/"}),
				Exclude: ss([]string{"*.txt", "*.rpm"}),
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"--include=testdata/",
				"--exclude=*.txt",
				"--exclude=*.rpm",
				"./",
				"jqhacker@example.com:/dev/null",
			},
		},
		{
			"example.com",
			".",
			Rsync{
				User:   "jqhacker",
				Source: "./",
				Target: "/dev/null",
				Filter: ss([]string{"+ testdata/", "- testdata/*.txt", "- testdata/*.rpm"}),
			},
			[]string{
				"rsync",
				"-az",
				"-e",
				"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
				"--filter=+ testdata/",
				"--filter=- testdata/*.txt",
				"--filter=- testdata/*.rpm",
				"./",
				"jqhacker@example.com:/dev/null",
			},
		},
	}
	for i, data := range testdata {
		c := data.rs.buildRsync(data.host, data.root)
		if len(c.Args) != len(data.exp) {
			t.Fatalf("Case %d: Expected %d, got %d", i, len(data.exp), len(c.Args))
		}
		for j := range c.Args {
			if c.Args[j] != data.exp[j] {
				t.Fatalf("Case %d:\nExpected:\n\t%s\nGot:\n\t%s", j, strings.Join(data.exp, " "), strings.Join(c.Args, " "))
			}
		}
	}
}

func abs(path string) string {
	s, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return s
}

func ss(args []string) drone.StringSlice {
	j, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}
	s := drone.StringSlice{}
	err = s.UnmarshalJSON(j)
	if err != nil {
		panic(err)
	}
	return s
}
