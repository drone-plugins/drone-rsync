package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
	"golang.org/x/crypto/ssh"
)

type Rsync struct {
	Hosts drone.StringSlice `json:"host"`
	host  string            `json:"-"`

	User      string   `json:"user"`
	Port      int      `json:"port"`
	Source    string   `json:"source"`
	Target    string   `json:"target"`
	Delete    bool     `json:"delete"`
	Recursive bool     `json:"recursive"`
	Exclude   string   `json:"exclude"`
	Commands  []string `json:"commands"`
}

func main() {
	w := new(drone.Workspace)
	v := new(Rsync)
	plugin.Param("workspace", w)
	plugin.Param("vargs", v)
	plugin.Parse()

	// write the rsa private key if provided
	if err := writeKey(w); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

		// sets the current host
		v.host = host

		// sync the files on the remote machine
		rs := buildRsync(v)
		rs.Dir = w.Path
		rs.Stderr = os.Stderr
		rs.Stdout = os.Stdout
		trace(rs)
		err := rs.Run()
		if err != nil {
			os.Exit(1)
			return
		}

		// continue if no commands
		if len(v.Commands) == 0 {
			continue
		}

		// execute commands on remote server (reboot instance, etc)
		if err := run(v, w.Keys); err != nil {
			os.Exit(1)
			return
		}
	}
}

// Build rsync command
func buildRsync(rs *Rsync) *exec.Cmd {

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

	// append files to exclude
	if len(rs.Exclude) != 0 {
		args = append(args, fmt.Sprintf("--exclude=%s", rs.Exclude))
	}

	args = append(args, rs.Source)
	args = append(args, fmt.Sprintf("%s@%s:%s", rs.User, rs.host, rs.Target))

	return exec.Command("rsync", args...)
}

// Run commands on the remote host
func run(rs *Rsync, keys *drone.Key) error {

	// join the host and port if necessary
	addr := net.JoinHostPort(
		rs.host,
		strconv.Itoa(rs.Port),
	)

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
