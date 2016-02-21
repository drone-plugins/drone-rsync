package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
	"golang.org/x/crypto/ssh"
)

type Rsync struct {
	Hosts     drone.StringSlice `json:"host"`
	User      string            `json:"user"`
	Port      int               `json:"port"`
	Source    string            `json:"source"`
	Target    string            `json:"target"`
	Delete    bool              `json:"delete"`
	Recursive bool              `json:"recursive"`
	Include   drone.StringSlice `json:"include"`
	Exclude   drone.StringSlice `json:"exclude"`
	Filter    drone.StringSlice `json:"filter"`
	Commands  []string          `json:"commands"`
}

var (
	buildCommit string
)

func main() {
	fmt.Printf("Drone Rsync Plugin built from %s\n", buildCommit)

	w := new(drone.Workspace)
	v := new(Rsync)
	plugin.Param("workspace", w)
	plugin.Param("vargs", v)
	if err := plugin.Parse(); err != nil {
		fmt.Println("Rsync: unable to parse invalid plugin input.")
		os.Exit(1)
	}
	if err := rsync(w, v); err != nil {
		fmt.Printf("Rsync: %s\n", err)
		os.Exit(1)
	}
}

func rsync(w *drone.Workspace, v *Rsync) error {
	// write the rsa private key if provided
	if err := writeKey(w); err != nil {
		return err
	}

	// default values
	if v.Port == 0 {
		v.Port = 22
	}
	if len(v.User) == 0 {
		v.User = "root"
	}
	if len(v.Source) == 0 {
		v.Source = "./"
	}

	// execute for each host
	for _, host := range v.Hosts.Slice() {
		// sync the files on the remote machine
		rs := v.buildRsync(host, w.Path)
		rs.Stderr = os.Stderr
		rs.Stdout = os.Stdout
		trace(rs)
		err := rs.Run()
		if err != nil {
			return err
		}

		// continue if no commands
		if len(v.Commands) == 0 {
			continue
		}

		// execute commands on remote server (reboot instance, etc)
		if err := v.run(w.Keys, host); err != nil {
			return err
		}
	}

	return nil
}

// Build rsync command
func (rs *Rsync) buildRsync(host, root string) *exec.Cmd {

	var args []string
	args = append(args, "-az")

	// append recursive flag
	if rs.Recursive {
		args = append(args, "-r")
	}

	// append delete flag
	if rs.Delete {
		args = append(args, "--del")
	}

	// append custom ssh parameters
	args = append(args, "-e")
	args = append(args, fmt.Sprintf("ssh -p %d -o UserKnownHostsFile=/dev/null -o LogLevel=quiet -o StrictHostKeyChecking=no", rs.Port))

	// append filtering rules
	for _, pattern := range rs.Include.Slice() {
		args = append(args, fmt.Sprintf("--include=%s", pattern))
	}
	for _, pattern := range rs.Exclude.Slice() {
		args = append(args, fmt.Sprintf("--exclude=%s", pattern))
	}
	for _, pattern := range rs.Filter.Slice() {
		args = append(args, fmt.Sprintf("--filter=%s", pattern))
	}

	args = append(args, rs.globSource(root)...)
	args = append(args, fmt.Sprintf("%s@%s:%s", rs.User, host, rs.Target))

	return exec.Command("rsync", args...)
}

// Run commands on the remote host
func (rs *Rsync) run(keys *drone.Key, host string) error {

	// join the host and port if necessary
	addr := net.JoinHostPort(host, strconv.Itoa(rs.Port))

	// trace command used for debugging in the build logs
	fmt.Printf("$ ssh %s@%s -p %d\n", rs.User, addr, rs.Port)

	signer, err := ssh.ParsePrivateKey([]byte(keys.Private))
	if err != nil {
		return fmt.Errorf("Error parsing private key. %s.", err)
	}

	config := &ssh.ClientConfig{
		User: rs.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("Error dialing server. %s.", err)
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Error starting ssh session. %s.", err)
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	return session.Run(strings.Join(rs.Commands, "\n"))
}

// globSource returns the names of all files matching the source pattern.
// If there are no matches or an error occurs, the original source string is
// returned.
//
// If the source path is not absolute the root path will be prepended to the
// source path prior to matching.
func (rs *Rsync) globSource(root string) []string {
	src := rs.Source
	if !path.IsAbs(rs.Source) {
		src = path.Join(root, rs.Source)
	}
	srcs, err := filepath.Glob(src)
	if err != nil || len(srcs) == 0 {
		return []string{rs.Source}
	}
	sep := fmt.Sprintf("%c", os.PathSeparator)
	if strings.HasSuffix(rs.Source, sep) {
		// Add back the trailing slash removed by path.Join()
		for i := range srcs {
			srcs[i] += sep
		}
	}
	return srcs
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}

// Writes the RSA private key
func writeKey(in *drone.Workspace) error {
	if in.Keys == nil || len(in.Keys.Private) == 0 {
		return nil
	}
	home := "/root"
	u, err := user.Current()
	if err == nil {
		home = u.HomeDir
	}
	sshpath := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshpath, 0700); err != nil {
		return err
	}
	confpath := filepath.Join(sshpath, "config")
	privpath := filepath.Join(sshpath, "id_rsa")
	ioutil.WriteFile(confpath, []byte("StrictHostKeyChecking no\n"), 0700)
	return ioutil.WriteFile(privpath, []byte(in.Keys.Private), 0600)
}
