// Copyright (c) 2014, Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tao

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"
	"syscall"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"

	"github.com/jlmucb/cloudproxy/tao/auth"
	"github.com/jlmucb/cloudproxy/util"
)

// A CoreOSConfig contains the details needed to start a new CoreOS VM.
type CoreOSConfig struct {
	Name       string
	ImageFile  string
	SSHPort    int
	Memory     int
	RulesPath  string
	SSHKeysCfg string
	SocketPath string
}

// A KVMCoreOSContainer is a simple wrapper for CoreOS running on KVM. It uses
// os/exec.Cmd to send commands to QEMU/KVM to start CoreOS then uses SSH to
// connect to CoreOS to start the LinuxHost there with a virtio-serial
// connection for its communication with the Tao running on Linux in the guest.
// This use of os/exec is to avoid having to rewrite or hook into libvirt for
// now.
type KvmCoreOSContainer struct {
	Cfg            *CoreOSConfig
	DockerHostFile string
	Args           []string
	QCmd           *exec.Cmd
}

// Kill sends a SIGKILL signal to a QEMU instance.
func (kcc *KvmCoreOSContainer) Kill() error {
	// Kill the qemu command directly.
	// TODO(tmroeder): rewrite this using qemu's communication/management
	// system; sending SIGKILL is definitely not the right way to do this.
	return kcc.QCmd.Process.Kill()
}

// Start starts a QEMU/KVM CoreOS container using the command line.
func (kcc *KvmCoreOSContainer) Start() error {
	// Create a temporary directory for the config drive.
	td, err := ioutil.TempDir("", "coreos")
	if err != nil {
		return err
	}
	// TODO(tmroeder): save this path and remove it on Stop/Kill.
	// defer os.RemoveAll(td)

	// Create a temporary directory for the LinuxHost Docker image.
	td_docker, err := ioutil.TempDir("", "coreos_docker")
	if err != nil {
		return err
	}
	// TODO(tmroeder): save this path and remove it on Stop/Kill.
	// defer os.RemoveAll(td_docker)
	linuxHostImage, err := ioutil.ReadFile(kcc.DockerHostFile)
	if err != nil {
		return err
	}
	linuxHostPath := path.Join(td_docker, "img")
	if err := ioutil.WriteFile(linuxHostPath, linuxHostImage, 0700); err != nil {
		return err
	}

	latestDir := path.Join(td, "openstack/latest")
	if err := os.MkdirAll(latestDir, 0700); err != nil {
		return err
	}

	cfg := kcc.Cfg
	userData := path.Join(latestDir, "user_data")
	if err := ioutil.WriteFile(userData, []byte(cfg.SSHKeysCfg), 0700); err != nil {
		return err
	}

	// Copy the rules into the mirrored filesystem for use by the Linux host
	// on CoreOS.
	rules, err := ioutil.ReadFile(cfg.RulesPath)
	if err != nil {
		return err
	}
	rulesFile := path.Join(latestDir, "rules")
	if err := ioutil.WriteFile(rulesFile, []byte(rules), 0700); err != nil {
		return err
	}

	qemuProg := "qemu-system-x86_64"
	qemuArgs := []string{"-name", cfg.Name,
		"-m", strconv.Itoa(cfg.Memory),
		"-machine", "accel=kvm:tcg",
		// Networking.
		"-net", "nic,vlan=0,model=virtio",
		"-net", "user,vlan=0,hostfwd=tcp::" + strconv.Itoa(cfg.SSHPort) + "-:22,hostname=" + cfg.Name,
		// Tao communications through virtio-serial. With this
		// configuration, QEMU waits for a server on cfg.SocketPath,
		// then connects to it.
		"-chardev", "socket,path=" + cfg.SocketPath + ",id=port0-char",
		"-device", "virtio-serial",
		"-device", "virtserialport,id=port1,name=tao,chardev=port0-char",
		// The CoreOS image to boot from.
		"-drive", "if=virtio,file=" + cfg.ImageFile,
		// A Plan9P filesystem for SSH configuration (and our rules).
		"-fsdev", "local,id=conf,security_model=none,readonly,path=" + td,
		"-device", "virtio-9p-pci,fsdev=conf,mount_tag=config-2",
		// Another Plan9P filesystem for the Docker files.
		"-fsdev", "local,id=tao,security_model=none,readonly,path=" + td_docker,
		"-device", "virtio-9p-pci,fsdev=tao,mount_tag=tao",
		// Machine config.
		"-cpu", "host",
		"-smp", "4",
		"-nographic"} // for now, we add -nographic explicitly.
	// TODO(tmroeder): append args later.
	//qemuArgs = append(qemuArgs, kcc.Args...)

	qemuCmd := exec.Command(qemuProg, qemuArgs...)
	// Don't connect QEMU/KVM to any of the current input/output channels,
	// since we'll connect over SSH.
	//qemuCmd.Stdin = os.Stdin
	//qemuCmd.Stdout = os.Stdout
	//qemuCmd.Stderr = os.Stderr
	kcc.QCmd = qemuCmd
	return kcc.QCmd.Start()
}

// Stop sends a SIGSTOP signal to a docker container.
func (kcc *KvmCoreOSContainer) Stop() error {
	// Stop the QEMU/KVM process with SIGSTOP.
	// TODO(tmroeder): rewrite this using qemu's communication/management
	// system; sending SIGSTOP is definitely not the right way to do this.
	return kcc.QCmd.Process.Signal(syscall.SIGSTOP)
}

// ID returns a numeric ID for this container. For now, this ID is 0.
func (kcc *KvmCoreOSContainer) ID() int {
	return 0
}

// A LinuxKvmCoreOSContainerFactory manages hosted programs started as QEMU/KVM
// instances over a given CoreOS image.
type LinuxKvmCoreOSContainerFactory struct {
	Cfg            *CoreOSConfig
	SocketPath     string
	DockerHostFile string
	NextSSHPort    int
	Mutex          sync.Mutex
}

// NewLinuxKvmCoreOSContainerFactory returns a new HostedProgramFactory that can
// create docker containers to wrap programs.
func NewLinuxKvmCoreOSContainerFactory(sockPath, dockerPath string, cfg *CoreOSConfig) HostedProgramFactory {
	return &LinuxKvmCoreOSContainerFactory{
		Cfg:            cfg,
		SocketPath:     sockPath,
		DockerHostFile: dockerPath,
		// The first SSH port is the port from the incoming config.
		NextSSHPort: cfg.SSHPort,
	}
}

// CloudConfigFromSSHKeys converts an ssh authorized-keys file into a format
// that can be used by CoreOS to authorize incoming SSH connections over the
// Plan9P-mounted filesystem it uses.
func CloudConfigFromSSHKeys(keysFile string) (string, error) {
	sshKeys := "#cloud-config\nssh_authorized_keys:"
	sshFile, err := os.Open(keysFile)
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(sshFile)
	for scanner.Scan() {
		sshKeys += "\n - " + scanner.Text()
	}

	return sshKeys, nil
}

// MakeSubprin computes the hash of a QEMU/KVM CoreOS image to get a
// subprincipal for authorization purposes.
func (ldcf *LinuxKvmCoreOSContainerFactory) MakeSubprin(id uint, image string) (auth.SubPrin, string, error) {
	var empty auth.SubPrin
	// TODO(tmroeder): the combination of TeeReader and ReadAll doesn't seem
	// to copy the entire image, so we're going to hash in place for now.
	// This needs to be fixed to copy the image so we can avoid a TOCTTOU
	// attack.
	b, err := ioutil.ReadFile(image)
	if err != nil {
		return empty, "", err
	}

	h := sha256.Sum256(b)
	subprin := FormatCoreOSSubprin(id, h[:])
	return subprin, image, nil
}

// FormatCoreOSSubprin produces a string that represents a subprincipal with the
// given ID and hash.
func FormatCoreOSSubprin(id uint, hash []byte) auth.SubPrin {
	var args []auth.Term
	if id != 0 {
		args = append(args, auth.Int(id))
	}
	args = append(args, auth.Bytes(hash))
	return auth.SubPrin{auth.PrinExt{Name: "CoreOS", Arg: args}}
}

func getRandomFileName(n int) string {
	// Get a random name for the socket.
	nameBytes := make([]byte, n)
	if _, err := rand.Read(nameBytes); err != nil {
		return ""
	}
	return hex.EncodeToString(nameBytes)
}

var nameLen = 10

// Launch launches a QEMU/KVM CoreOS instance, connects to it with SSH to start
// the LinuxHost on it, and returns the socket connection to that host.
func (ldkcf *LinuxKvmCoreOSContainerFactory) Launch(imagePath string, args []string) (io.ReadWriteCloser, HostedProgram, error) {
	// Build the new Config and start it. Make sure it has a random name so
	// it doesn't conflict with other virtual machines. Note that we need to
	// assign fresh local SSH ports for each new virtual machine, hence the
	// mutex and increment operation.
	sockName := getRandomFileName(nameLen)
	sockPath := path.Join(ldkcf.SocketPath, sockName)

	ldkcf.Mutex.Lock()
	sshPort := ldkcf.NextSSHPort
	ldkcf.NextSSHPort += 1
	ldkcf.Mutex.Unlock()

	cfg := &CoreOSConfig{
		Name:       getRandomFileName(nameLen),
		ImageFile:  imagePath,
		SSHPort:    sshPort,
		Memory:     ldkcf.Cfg.Memory,
		RulesPath:  ldkcf.Cfg.RulesPath,
		SSHKeysCfg: ldkcf.Cfg.SSHKeysCfg,
		SocketPath: sockPath,
	}

	// Create a new docker image from the filesystem tarball, and use it to
	// build a container and launch it.
	kcc := &KvmCoreOSContainer{
		Cfg:            cfg,
		DockerHostFile: ldkcf.DockerHostFile,
		Args:           args,
	}

	// Create the listening server before starting the connection. This lets
	// QEMU start right away. See the comments in Start, above, for why this
	// is.
	rwc := util.NewUnixSingleReadWriteCloser(cfg.SocketPath)
	if err := kcc.Start(); err != nil {
		return nil, nil, err
	}

	// We need some way to wait for the socket to open before we can connect
	// to it and return the ReadWriteCloser for communication. Also we need
	// to connect by SSH to the instance once it comes up properly. For now,
	// we just wait for a timeout before trying to connect and listen.
	tc := time.After(10 * time.Second)

	// TODO(tmroeder): for now, this program expects to find SSH_AUTH_SOCK
	// in its environment so it can connect to the local ssh-agent for ssh
	// keys. In the future, this will need to be configured better.
	agentPath := os.ExpandEnv("$SSH_AUTH_SOCK")
	if agentPath == "" {
		return nil, nil, fmt.Errorf("couldn't find agent socket in the environment.\n")
	}

	// TODO(tmroeder): the right way to do this is to first set up a docker
	// container for LinuxHost so it can run directly on CoreOS. Then add it
	// to the config directory and make it part of the name. Then start it
	// with a docker command and hook up /dev/tao directly to it. For
	// initial testing, though, we'll try to set up the virtual machine and
	// run a simple command on it, then fail.
	log.Printf("Found agent socket path '%s'\n", agentPath)
	agentSock, err := net.Dial("unix", agentPath)
	if err != nil {
		return nil, nil, fmt.Errorf("Couldn't connect to agent socket '%s': %s\n", agentPath, err)
	}
	ag := agent.NewClient(agentSock)

	// Set up an ssh client config to use to connect to CoreOS.
	conf := &ssh.ClientConfig{
		// The CoreOS user for the SSH keys is currently always 'core'.
		User: "core",
		Auth: []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)},
	}

	log.Println("Waiting for at most 10 seconds before trying to connect")
	<-tc

	hostPort := net.JoinHostPort("localhost", strconv.Itoa(sshPort))
	client, err := ssh.Dial("tcp", hostPort, conf)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't dial '%s'\n", hostPort)
	}

	// We need to run a set of commands to set up the LinuxHost as a docker
	// container on the remote system.
	// Mount the filesystem.
	mount, err := client.NewSession()
	mount.Stdout = os.Stdout
	mount.Stderr = os.Stderr
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't establish a mount session on SSH: %s\n", err)
	}
	if err := mount.Run("sudo mkdir /media/tao && sudo mount -t 9p -o trans=virtio,version=9p2000.L tao /media/tao"); err != nil {
		return nil, nil, fmt.Errorf("couldn't mount the tao filesystem on the guest: %s\n", err)
	}
	mount.Close()

	// Build the container.
	build, err := client.NewSession()
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't establish a build session on SSH: %s\n", err)
	}
	if err := build.Run("sudo cat /media/tao/img | sudo docker build -t tao -"); err != nil {
		return nil, nil, fmt.Errorf("couldn't build the tao linux_host docker image on the guest: %s\n", err)
	}
	build.Close()

	// Create the linux_host on the container.
	//	create, err := client.NewSession()
	//	create.Stdout = os.Stdout
	//	create.Stderr = os.Stderr
	//	if err != nil {
	//		return nil, nil, fmt.Errorf("couldn't establish a create session on SSH: %s\n", err)
	//	}
	//	if err := create.Run("sudo docker run --privileged=true -v /dev/virtio-ports/tao:/tao -v /media/configvirtfs/openstack/latest/rules:/home/tmroeder/src/github.com/jlmucb/cloudproxy/test/rules tao ./bin/linux_host --create --stacked"); err != nil {
	//		return nil, nil, fmt.Errorf("couldn't start linux_host on the guest: %s\n", err)
	//	}
	//	create.Close()

	// Start the linux_host on the container.
	start, err := client.NewSession()
	start.Stdout = os.Stdout
	start.Stderr = os.Stderr
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't establish a start session on SSH: %s\n", err)
	}
	if err := start.Run("sudo docker run -d --name=\"linux_host\" --privileged=true -v /dev/virtio-ports/tao:/tao -v /media/configvirtfs/openstack/latest/rules:/home/tmroeder/src/github.com/jlmucb/cloudproxy/test/rules tao ./bin/linux_host --service --stacked"); err != nil {
		return nil, nil, fmt.Errorf("couldn't start linux_host on the guest: %s\n", err)
	}
	start.Close()

	return rwc, kcc, nil
}
