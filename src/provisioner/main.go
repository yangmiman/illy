package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"provisioner/cert"
	"provisioner/fs"
	"provisioner/provisioner"
	"provisioner/provisioner/commands"
	"strconv"
	"syscall"
	"time"
)

var (
	provisionScriptPath = "/var/pcfdev/run"
	timeoutInSeconds    = "3600"
	distro              = "pcf"
)

func main() {
	checkArgCount()

	provisionTimeout, err := strconv.Atoi(timeoutInSeconds)
	if err != nil {
		fmt.Printf("Error: %s.", err)
		os.Exit(1)
	}

	silentCommandRunner := &provisioner.ConcreteCmdRunner{
		Stdout:  ioutil.Discard,
		Stderr:  ioutil.Discard,
		Timeout: time.Duration(provisionTimeout) * time.Second,
	}
	p := &provisioner.Provisioner{
		Cert: &cert.Cert{},
		CmdRunner: &provisioner.ConcreteCmdRunner{
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Timeout: time.Duration(provisionTimeout) * time.Second,
		},
		FS:       &fs.FS{},
		Commands: buildCommands(silentCommandRunner),

		Distro: distro,
	}

	if err := p.Provision(provisionScriptPath, os.Args[1:]...); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus())
				} else {
					os.Exit(1)
				}
			}
		case *provisioner.TimeoutError:
			fmt.Printf("Timed out after %s seconds.\n", timeoutInSeconds)
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}

func checkArgCount() {
	if len(os.Args) < 6 {
		fmt.Println("Need 5 arguments, Usage: ./provision <domain> <ip> <services> <docker_registries> <provider>")
		os.Exit(1)
	}
}

func buildCommands(commandRunner provisioner.CmdRunner) []provisioner.Command {
	providerAgnostic := []provisioner.Command{
		&commands.DisableUAAHSTS{
			WebXMLPath: "/var/vcap/packages/uaa/tomcat/conf/web.xml",
		},
		&commands.ConfigureDnsmasq{
			Domain:     os.Args[1],
			ExternalIP: os.Args[2],
			FS:         &fs.FS{},
			CmdRunner:  commandRunner,
		},
		&commands.ConfigureGardenDNS{
			FS:        &fs.FS{},
			CmdRunner: commandRunner,
		},
		&commands.SetupApi{
			CmdRunner: commandRunner,
			FS:        &fs.FS{},
		},
		&commands.ReplaceDomain{
			CmdRunner: commandRunner,
			FS:        &fs.FS{},
			NewDomain: os.Args[1],
		},
		&commands.SetupCFDot{
			CmdRunner: commandRunner,
			FS:        &fs.FS{},
		},
	}

	const (
		httpPort      = "80"
		httpsPort     = "443"
		sshPort       = "22"
		sshProxyPort  = "2222"
		tcpPortLower  = 61001
		tcpPortHigher = 61100
	)

	forAwsProvider := []provisioner.Command{
		&commands.CloseAllPorts{
			CmdRunner: commandRunner,
		},
		&commands.OpenPort{
			CmdRunner: commandRunner,
			Port:      httpPort,
		},
		&commands.OpenPort{
			CmdRunner: commandRunner,
			Port:      httpsPort,
		},
		&commands.OpenPort{
			CmdRunner: commandRunner,
			Port:      sshPort,
		},
		&commands.OpenPort{
			CmdRunner: commandRunner,
			Port:      sshProxyPort,
		},
	}

	for p := tcpPortLower; p <= tcpPortHigher; p++ {
		forAwsProvider = append(forAwsProvider, &commands.OpenPort{
			CmdRunner: commandRunner,
			Port:      strconv.Itoa(p),
		})
	}

	if isAwsProvisioner() {
		return append(providerAgnostic, forAwsProvider...)
	} else {
		return providerAgnostic
	}
}

func isAwsProvisioner() bool {
	return os.Args[5] == "aws"
}
