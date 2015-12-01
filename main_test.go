package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/drone/drone-go/drone"
)

var testdata = []struct {
	host string
	root string
	rs   Rsync
	exp  []string
	err  bool
}{
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: "testdata/*.txt",
			Target: "/home/drone/testdata",
		},
		[]string{
			"rsync",
			"-az",
			"-e",
			"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
			"testdata/bar.txt",
			"testdata/foo.txt",
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: "testdata/foo*",
			Target: "/home/drone/testdata",
		},
		[]string{
			"rsync",
			"-az",
			"-e",
			"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
			"testdata/foo",
			"testdata/foo.txt",
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: "testdata/bar-*.x86_64.rpm",
			Target: "/home/drone/testdata",
		},
		[]string{
			"rsync",
			"-az",
			"-e",
			"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
			"testdata/bar-0.0.30+g4cdc188-1.x86_64.rpm",
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: "notthere/*.txt",
			Target: "/home/drone/testdata",
		},
		[]string{
			"rsync",
			"-az",
			"-e",
			"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
			"notthere/*.txt",
			"drone@localhost:/home/drone/testdata",
		},
		true,
	},
	{
		"localhost",
		".",
		Rsync{
			User:      "drone",
			Source:    "testdata/",
			Target:    "/home/drone/testdata",
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
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: "testdata/foo.txt",
			Target: "/home/drone/testdata",
		},
		[]string{
			"rsync",
			"-az",
			"-e",
			"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
			"testdata/foo.txt",
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: "testdata/bar.txt",
			Target: "/home/drone/testdata",
		},
		[]string{
			"rsync",
			"-az",
			"-e",
			"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
			"testdata/bar.txt",
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: abs("testdata/bar.txt"),
			Target: "/home/drone/testdata",
		},
		[]string{
			"rsync",
			"-az",
			"-e",
			"ssh -p 0 -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no",
			abs("testdata/bar.txt"),
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:    "drone",
			Target:  "/home/drone/testdata",
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
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
	{
		"localhost",
		".",
		Rsync{
			User:   "drone",
			Source: "./",
			Target: "/home/drone/testdata",
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
			"drone@localhost:/home/drone/testdata",
		},
		false,
	},
}

func TestRsync(t *testing.T) {
	if len(os.Getenv("DRONE")) == 0 {
		t.Skip("Skipping unless run under Drone CI")
	}
	w := drone.Workspace{
		Keys: readKey("testdata/id_rsa"),
		Path: os.Getenv("DRONE_DIR"),
	}
	for i, data := range testdata {
		v := data.rs
		v.Hosts = ss([]string{data.host})
		err := rsync(&w, &v)
		if err != nil {
			if !data.err {
				t.Errorf("Case %d: %s", i, err)
			}
		} else {
			if data.err {
				t.Errorf("Case %d: %s", i, err)
			}
		}
	}
}

func TestBuildRsync(t *testing.T) {
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

func readKey(priv string) *drone.Key {
	pubKey, err := ioutil.ReadFile(fmt.Sprintf("%s.pub", priv))
	if err != nil {
		panic(err)
	}
	privKey, err := ioutil.ReadFile(priv)
	if err != nil {
		panic(err)
	}
	return &drone.Key{
		Public:  string(pubKey[:]),
		Private: string(privKey[:]),
	}
}
