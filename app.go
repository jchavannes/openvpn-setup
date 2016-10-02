package main

import (
	"os/exec"
	"os"
	"fmt"
	"bufio"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		outputHelp()
		os.Exit(0)
	}
	switch args[0] {
	case "setup-server":
		println("Setting up server...")
		if isOpenVPNInstalled() {
			println("OpenVPN installed.")
		} else {
			println("Installing OpenVPN...")
			installOpenVPN()
		}
		status()
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
	print("... OpenVPN status ...\n")
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
