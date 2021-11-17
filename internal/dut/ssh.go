// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dut controls devices under test (DUT) for ONDATRA tests.
package dut

import (
	"bytes"
	"fmt"
	"io/ioutil"
	log "github.com/golang/glog"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Opts struct {	
	debug              bool
}

func (o *Opts) Debug() bool {
	return o.debug
}

type SshClient struct {
	opts     *Opts
	location string
	username string
	config   *ssh.ClientConfig
	client   *ssh.Client
}

type DutInterface struct {
	Name     string
	MacAddr  string
	Ipv4Addr string
	Ipv6Addr string
}

func NewSshClient(opts *Opts, location string, username string) (*SshClient, error) {
	log.Infof("Creating SSH client for server %s ...\n", location)
	sshConfig := ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			return nil, nil
		})},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	log.Info("Dialing SSH ...")
	client, err := ssh.Dial("tcp", location, &sshConfig)
	if err != nil {
		return nil, fmt.Errorf("could not dial SSH server %s: %v", location, err)
	}
	return &SshClient{
		opts:     opts,
		location: location,
		username: username,
		config:   &sshConfig,
		client:   client,
	}, nil
}

func (c *SshClient) Close() {
	log.Infof("Closing SSH connection with %s\n", c.location)
	c.client.Close()
}

func (c *SshClient) Exec(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("could not create ssh session: %v", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if c.opts.Debug() {
		log.Infof("SSH INPUT: %s\n", cmd)
	}
	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("could not execute '%s': %v", cmd, err)
	}

	out := b.String()
	if c.opts.Debug() {
		log.Infof("SSH OUTPUT: %s\n", out)
	}
	return out, nil
}

func (c *SshClient) ExecMultiple(cmds []string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("could not create ssh session: %v", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	for _, cmd := range cmds {
		if c.opts.Debug() {
			log.Infof("SSH INPUT: %s\n", cmd)
		}
		if err := session.Run(cmd); err != nil {
			return "", fmt.Errorf("could not execute '%s': %v", cmd, err)
		}
	}

	out := b.String()
	if c.opts.Debug() {
		log.Infof("SSH OUTPUT: %s\n", out)
	}
	return out, nil
}

func (c *SshClient) PushDutConfigFile(location string) (string, error) {
	log.Infof("Reading DUT config %s ...", location)
	bytes, err := ioutil.ReadFile(location)
	if err != nil {
		return "", fmt.Errorf("could not read DUT config %s: %v", location, err)
	}

	return c.Exec("enable\nconfig terminal\n" + string(bytes))
}

func (c *SshClient) ResetDutConfig() (string, error) {
	return c.Exec("enable\nconfig terminal\nreset")
}

func (c *SshClient) GetInterface(name string) (*DutInterface, error) {
	ifc := DutInterface{}
	out, err := c.Exec("show interface " + name)
	if err != nil {
		return nil, err
	}

	ifc.Name = name
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "Hardware is Ethernet") {
			for _, word := range strings.Split(line, " ") {
				if strings.Contains(word, ".") {
					ifc.MacAddr = fmt.Sprintf(
						"%s:%s:%s:%s:%s:%s",
						word[0:2], word[2:4], word[5:7], word[7:9], word[10:12],
						word[12:14],
					)
				}
			}
		} else if strings.Contains(line, "Internet address") {
			for _, word := range strings.Split(line, " ") {
				if strings.Contains(word, ".") {
					ifc.Ipv4Addr = strings.Split(word, "/")[0]
				}
			}
		}
	}

	return &ifc, nil
}
