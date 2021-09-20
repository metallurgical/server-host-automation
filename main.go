package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var domain, projectRoot, gitEndpoint, whichWebServer, phpVersion string

func main() {
	fmt.Println("[ Server host automation tool by @metallurgical(https://github.com/metallurgical) ]")
	askForInput()

	if projectRoot != "" && gitEndpoint != "" {
		cmd := exec.Command("git", "clone", gitEndpoint, projectRoot)
		cmd.Run();
	}

	if whichWebServer == "2" {
		createNginxVhost();
	} else {
		createApacheVhost();
	}
}

func askForInput() {
	fmt.Print("1) What is the domain name (without http/https) ? : ")
	fmt.Scanln(&domain)

	fmt.Print("2) Where to clone the git repo (Full path to base folder of project) ? : ")
	fmt.Scanln(&projectRoot)

	fmt.Print("3) Full URL of git endpoint (will be automatically cloned into folder set in no 2) ? : ")
	fmt.Scanln(&gitEndpoint)

	fmt.Print("4) Which web server? (Apache - 1, Nginx - 2, key in 1 or 2) ? : ")
	fmt.Scanln(&whichWebServer)

	fmt.Print("5) Write down current server PHP version? (Leave empty if using apache) ? : ")
	fmt.Scanln(&phpVersion)
}

func createNginxVhost() {
	_, err := os.Stat("/etc/nginx/sites-available")
	if err != nil {
		return
	}
	// Get/download the file from source
	cmd := exec.Command("wget", "https://raw.githubusercontent.com/metallurgical/server-host-automation/master/default-nginx-host.conf", "-P", "/tmp");
	// Copy the file from source into nginx sites-available folder
	var vhostFileName = domain + ".conf"
	var domainPath = "/etc/nginx/site-available/" + vhostFileName
	cmd = exec.Command("cp", "/tmp/" + vhostFileName, domainPath)
	cmd.Run()

	// Replace document root full path into new project directory path
	replaceContent(domainPath, "$documentRoot", projectRoot + "/public");
	// Replace matching server name with new the exact domain name
	replaceContent(domainPath, "$serverName", domain);
	// Replace php fpm socket path
	replaceContent(domainPath, "$phpFpmSocket", "unix:/var/run/php/php " + phpVersion + "-fpm.sock");

	// Once successfully created into sites-available, create symlink to that file
	cmd = exec.Command("ln", "-s", domainPath, "/etc/nginx/sites-enabled/")
	cmd.Run()
}

func createApacheVhost() {

}

func replaceContent(source string, search string, replace string) {
	input, err := ioutil.ReadFile(source)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	output := bytes.Replace(input, []byte(search), []byte(replace), -1)
	if err = ioutil.WriteFile(source, output, 0666); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
