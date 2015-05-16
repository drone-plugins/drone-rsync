package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/drone/drone-plugin-go/plugin"
)

var rsyncCommand = `rsync -az {{ if .Recursive }}-r {{ end }} {{ if .Delete }}--del {{ end }} -e "ssh -p {{ .Port }} -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no" --exclude={{ .Exclude }} {{ .Source }} {{ .User }}@{{ .Host }}:{{ .Target }}`

type Rsync struct {
	User      string `json:"user"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Delete    bool   `json:"delete"`
	Recursive bool   `json:"recursive"`
	Exclude   string `json:"exclude"`
}

func main() {
	c := new(plugin.Clone)
	v := new(Rsync)
	plugin.Param("clone", c)
	plugin.Param("vargs", v)
	plugin.Parse()

	// write the rsa private key if provided
	if err := writeKey(c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// create the rsync command
	rs := buildRsync(v)
	rs.Dir = c.Dir
	rs.Stderr = os.Stderr
	rs.Stdout = os.Stdout
	trace(rs)
	err := rs.Run()
	if err != nil {
		os.Exit(1)
		return
	}

	// and execute

	// create remote command script

	// and execute
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}

// Build rsync command
func buildRsync(rs *Rsync) *exec.Cmd {
	return nil
}

// Writes the RSA private key
func writeKey(in *plugin.Clone) error {
	if len(in.Keypair.Private) == 0 {
		return nil
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	sshpath := filepath.Join(u.HomeDir, ".ssh")
	if err := os.MkdirAll(sshpath, 0700); err != nil {
		return err
	}
	confpath := filepath.Join(sshpath, "config")
	privpath := filepath.Join(sshpath, "id_rsa")
	ioutil.WriteFile(confpath, []byte("StrictHostKeyChecking no\n"), 0700)
	return ioutil.WriteFile(privpath, []byte(in.Keypair.Private), 0600)
}
