package main

import (
	"os/exec"
	"os"
	"fmt"
	"bufio"
	"flag"
	"log"
)

type OpenVPNConfig struct {
	KEY_COUNTRY  string
	KEY_PROVINCE string
	KEY_CITY     string
	KEY_ORG      string
	KEY_EMAIL    string
	KEY_OU       string
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		outputHelp()
		os.Exit(0)
	}
	switch args[0] {
	case "setup-server":
		f1 := flag.NewFlagSet("f1", flag.ContinueOnError)
		var configFilename string
		f1.StringVar(&configFilename, "c", "", "Config file name.")
		f1.Parse(args[1:])

		if len(configFilename) == 0 {
			log.Fatal("Must specify config filename (-c).")
		}



		println("Setting up server...")
		if isOpenVPNInstalled() {
			println("OpenVPN installed.")
		} else {
			println("Installing OpenVPN...")
			installOpenVPN()
		}
		if isEasyRSASetup() {
			println("easy-rsa setup.")
		} else {
			if isEasyRSAInstalled() {
				println("easy-rsa already installed.")
			} else {
				println("Installing easy-rsa...")
				installEasyRSA()
			}
			println("Setting up easy-rsa...")
			setupEasyRSA()
			println("Initializing PKI...")
			initializePKI()
		}
	case "status":
		status()
	case "test":
		println("test...")
	default:
		print("Unknown command.\n\n")
		outputHelp()
	}
}

func installOpenVPN() {
	streamCommand("sudo", "apt-get", "install", "-y", "openvpn")
}

func status() {
	print("--== OpenVPN status ==--\n")
	fmt.Printf("Installed: %t\n", isOpenVPNInstalled())
	fmt.Printf("Configured as server: %t\n", isConfiguredAsServer())
	fmt.Printf("Configured as client: %t\n", isConfiguredAsClient())
}

func isOpenVPNInstalled() bool {
	err := exec.Command("which", "openvpn").Run()
	return err == nil
}

func isConfiguredAsServer() bool {
	err := exec.Command("test", "-f", "/etc/openvpn/server.conf").Run()
	return err == nil
}

func isConfiguredAsClient() bool {
	err := exec.Command("test", "-f", "/etc/openvpn/client.conf").Run()
	return err == nil
}

func isEasyRSASetup() bool {
	err := exec.Command("test", "-d", "/etc/openvpn/easy-rsa").Run()
	return err == nil
}

func setupEasyRSA() {
	streamCommand("bash", "-c", "cd /etc/openvpn && make-cadir easy-rsa")
}

func isEasyRSAInstalled() bool {
	err := exec.Command("which", "make-cadir").Run()
	return err == nil
}

func installEasyRSA() {
	streamCommand("sudo", "apt-get", "install", "-y", "easy-rsa")
}

func initializePKI() {
	exec.Command("ln", "-s", "/etc/openvpn/easy-rsa/openssl-1.0.0.cnf", "/etc/openvpn/easy-rsa/openssl.conf")
	streamCommand("bash", "-c", "cd /etc/openvpn/easy-rsa && source ./vars && ./clean-all")
	streamCommand("bash", "-c", "cd /etc/openvpn/easy-rsa && source ./vars && ./build-ca --batch")
}

func outputHelp() {
	print(
		"Available commands:\n",
		" - setup-server\n",
		" - status\n",
	)
}

func streamCommand(cmdName string, cmdArgs ...string) {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	cmdErrReader, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		os.Exit(1)
	}

	errScanner := bufio.NewScanner(cmdErrReader)
	go func() {
		for errScanner.Scan() {
			fmt.Printf("%s\n", errScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		os.Exit(1)
	}
}
