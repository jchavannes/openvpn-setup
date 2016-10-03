package main

import (
	"os/exec"
	"os"
	"fmt"
	"bufio"
	"flag"
	"log"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"math/rand"
	"strconv"
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
		println("Setting up server...")
		if isOpenVPNInstalled() {
			println("OpenVPN already installed.")
		} else {
			println("Installing OpenVPN...")
			installOpenVPN()
		}
		if isEasyRSASetup() {
			println("easy-rsa already setup.")
		} else {
			if isEasyRSAInstalled() {
				println("easy-rsa already installed.")
			} else {
				println("Installing easy-rsa...")
				installEasyRSA()
			}
			println("Setting up easy-rsa...")
			setupEasyRSA()
		}
		setupCustomVars()
		if isPKIInitialized() {
			println("PKI already initialized.")
		} else {
			println("Initializing PKI...")
			initializePKI()
		}
		if areOpenVPNKeysSetup() {
			println("OpenVPN keys already in right location.")
		} else {
			println("Copying keys to right location...")
			copyKeysToOpenVPNRoot()
		}
		if isConfiguredAsServer() {
			println("Already setup as server.")
		} else {
			println("Setting up server config...")
			setupServerConfig()
		}
	case "status":
		status()
	case "client":
		if len(args) < 3 {
			log.Fatal("Must specify client name (-n).")
		}

		f1 := flag.NewFlagSet("f1", flag.ContinueOnError)

		var name string
		f1.StringVar(&name, "n", "", "Client name.")
		f1.Parse(args[1:])

		if len(name) == 0 {
			log.Fatal("Must specify client name (-n).")
		}

		if doesClientExist(name) {
			println("Client already exists.")
		} else {
			print("Creating client...")
			createClient(name)
		}
		outputClient(name)
	case "test":
		println("test...")
	default:
		print("Unknown command.\n\n")
		outputHelp()
	}
}

func getServerConfig() *OpenVPNConfig {
	args := os.Args[1:]
	if len(args) < 3 {
		log.Fatal("Must specify config filename (-c).")
	}

	f1 := flag.NewFlagSet("f1", flag.ContinueOnError)

	var configFilename string
	f1.StringVar(&configFilename, "c", "", "Config file name.")
	f1.Parse(args[1:])

	if len(configFilename) == 0 {
		log.Fatal("Must specify config filename (-c).")
	}

	file, err := ioutil.ReadFile(configFilename)
	check(err)

	var openVPNConfig OpenVPNConfig
	err = json.Unmarshal(file, &openVPNConfig)
	check(err)

	return &openVPNConfig
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

func isPKIInitialized() bool {
	err := exec.Command("test", "-f", "/etc/openvpn/easy-rsa/keys/server.crt").Run()
	return err == nil
}

func initializePKI() {
	err := exec.Command("sed", "-i", "s/^\\(subjectAltName=\\)/# \\1/g", "/etc/openvpn/easy-rsa/openssl-1.0.0.cnf").Run()
	check(err)

	exec.Command("ln", "-s", "/etc/openvpn/easy-rsa/openssl-1.0.0.cnf", "/etc/openvpn/easy-rsa/openssl.conf")

	steps := []string{
		"cd /etc/openvpn/easy-rsa",
		"source ./vars > /dev/null",
		"source ./vars-custom",
		"export KEY_CN='server'",
		"./clean-all",
		"./build-ca --batch",
		"./build-key-server --batch server",
		"./build-dh",
	}
	streamCommand("bash", "-c", strings.Join(steps, " && "))
}

func setupCustomVars() {
	openVPNConfig := getServerConfig()
	value := reflect.ValueOf(*openVPNConfig)

	var exports []string
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		str := field.String()

		if field.Type().Kind() != reflect.String || len(str) == 0 {
			continue
		}

		exports = append(exports, "export " + value.Type().Field(i).Name + "=\"" + str + "\"")
	}

	ioutil.WriteFile("/etc/openvpn/easy-rsa/vars-custom", []byte(strings.Join(exports, "\n")), 0644)
}

func doesClientExist(name string) bool {
	err := exec.Command("test", "-f", "/etc/openvpn/easy-rsa/keys/" + name + ".crt").Run()
	return err == nil
}

func createClient(name string) {
	if !isPKIInitialized() {
		log.Fatal("PKI not initialized, run setup-server first")
	}

	steps := []string{
		"cd /etc/openvpn/easy-rsa",
		"source ./vars > /dev/null",
		"source ./vars-custom",
		"export KEY_CN='" + name + "'",
		"./build-key --batch " + name,
	}
	streamCommand("bash", "-c", strings.Join(steps, " && "))
}

func outputClient(name string) {
	out := func(filename string) {
		ca, err := ioutil.ReadFile(filename)
		check(err)
		fmt.Printf("%s:\n%s\n", filename, ca)
	}

	out("/etc/openvpn/ca.crt")
	out("/etc/openvpn/easy-rsa/keys/" + name + ".key")
	out("/etc/openvpn/easy-rsa/keys/" + name + ".crt")
}

func areOpenVPNKeysSetup() bool {
	err := exec.Command("test", "-f", "/etc/openvpn/server.key").Run()
	return err == nil
}

func copyKeysToOpenVPNRoot() {
	if !isPKIInitialized() {
		log.Fatal("PKI not initialized, run setup-server first")
	}

	steps := []string{
		"cd /etc/openvpn/easy-rsa/keys",
		"cp ca.crt ca.key dh2048.pem server.crt server.key /etc/openvpn",
	}
	streamCommand("bash", "-c", strings.Join(steps, " && "))
}

func setupServerConfig() {
	err := exec.Command("bash", "-c", "gunzip -c /usr/share/doc/openvpn/examples/sample-config-files/server.conf.gz > /etc/openvpn/server.conf").Run()
	check(err)
	err = exec.Command("sed", "-i", "s/dh dh1024\\.pem/dh dh2048.pem/g", "/etc/openvpn/server.conf").Run()
	check(err)
	err = exec.Command("sed", "-i", "s/^\\(server 10.\\)8\\(.0.0 255.255.255.0\\)/\\1" + strconv.Itoa(rand.Intn(100)) + "\\2/g", "/etc/openvpn/server.conf").Run()
	check(err)
}

func outputHelp() {
	print(
		"Available commands:\n",
		" - client\n",
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

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
