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
	return execIgnore("ps ax -o args | grep " + name + " | grep -v grep")
}

func execIgnore(command string) string {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}
