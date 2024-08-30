package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type VPN struct {
	Name     string
	Provider string
}

const defaultTemplate = `{{.Provider}} - {{.Name}}`
const defaultIfNotVPN = "none"

func main() {
	tmplString := flag.String("template", defaultTemplate, "The template to use")
	ifNotVPN := flag.String("ifNotVPN", defaultIfNotVPN, "The text to display if not connected to a VPN")
	flag.Parse()

	tmpl, err := template.New("vpn").Parse(*tmplString)
	if err != nil {
		err = fmt.Errorf("Error when parsing template: %s", err)
		panic(err)
	}

	out := findVPN()
	if out.Name == "" {
		fmt.Print(*ifNotVPN)
		return
	}
	tmpl.Execute(os.Stdout, out)
}

func findVPN() VPN {
	out := findOpenVPN()
	if out.Name != "" {
		return out
	}

	out = findFortiVPN()
	if out.Name != "" {
		return out
	}

	out = findUniVPN()
	if out.Name != "" {
		return out
	}

	return VPN{}
}

func findOpenVPN() VPN {
	out := findFirstProcess("openvpn")
	if out == "" {
		return VPN{}
	}

	// Find the config file
	split := strings.Split(out, " ")
	name := ""
	for i, s := range split {
		if s == "--config" && i < len(split)-1 {
			name = split[i+1]
			break
		}
	}

	if name == "" {
		return VPN{}
	}

	// Remove the quotes
	strings.Replace(name, "\"", "", -1)

	// Remove the path
	split = strings.Split(name, "/")
	name = split[len(split)-1]

	// Remove the extension
	split = strings.Split(name, ".")
	name = split[0]

	return VPN{
		Name:     name,
		Provider: "OpenVPN",
	}
}

func findFortiVPN() VPN {
	out := execIgnore("forticlient vpn status")

	dataMap := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		split := strings.Split(line, ":")
		if len(split) < 2 {
			continue
		}
		dataMap[strings.TrimSpace(split[0])] = strings.TrimSpace(split[1])
	}

	if dataMap["Status"] != "Connected" {
		return VPN{}
	}

	name := dataMap["VPN name"]

	return VPN{
		Name:     name,
		Provider: "FortiVPN",
	}
}

const uniVPNLogPath = "~/UniVPN/log/"

func findUniVPN() VPN {
	out := findFirstProcess("univpn")
	if out == "" {
		return VPN{}
	}

	// files, err := os.ReadDir(uniVPNLogPath)
	// if err != nil {
	// 	return VPN{}
	// }
	//
	// var lastFile os.DirEntry
	// var lastFileModTime time.Time
	// for _, file := range files {
	// 	if file.Type().IsDir() {
	// 		continue
	// 	}
	//
	// 	fileInfo, err := file.Info()
	// 	if err != nil {
	// 		continue
	// 	}
	//
	// 	if lastFile == nil || fileInfo.ModTime().After(lastFileModTime) {
	// 		lastFile = file
	// 		lastFileModTime = fileInfo.ModTime()
	// 	}
	// }
	//
	// if lastFile == nil {
	// 	return VPN{}
	// }
	//
	// file, err := os.Open(uniVPNLogPath + lastFile.Name())

	return VPN{
		Name:     "Unknown",
		Provider: "UniVPN",
	}
}

func findFirstProcess(name string) string {
	out := findProcess(name)
	if out == "" {
		return ""
	}
	return out[:strings.Index(out, "\n")]
}

func findProcess(name string) string {
	if name == "" {
		return ""
	}
	return execIgnore("ps ax -o args | grep -i " + name + " | grep -v grep")
}

func execIgnore(command string) string {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
