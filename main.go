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
	fmt.Println("[ Server host automation tool by @metallurgical(https://github.com/metallurgical)]")
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
}

func createNginxVhost() {
	_, err := os.Stat("/etc/nginx/sites-available")
	if err != nil {
		return
	}
	// Get/download the file from source
	cmdWget := exec.Command("wget", "https://raw.githubusercontent.com/metallurgical/server-host-automation/master/default-nginx-host.conf", "-P", "/tmp");
	cmdWget.Run();
	// Copy the file from source into nginx sites-available folder
	var vhostFileName = domain + ".conf"
	var domainPath = "/etc/nginx/site-available/" + vhostFileName
	cmdCp := exec.Command("cp", "/tmp/" + vhostFileName, domainPath)
	cmdCp.Run();

	// Replace document root full path into new project directory path
	replaceContent(domainPath, "[documentRoot]", projectRoot + "/public");
	// Replace matching server name with new the exact domain name
	replaceContent(domainPath, "[serverName]", domain);
	// Get the full path of php-fpm socket. This will return the output
	// something similar to this eg: "listen = /run/php/php7.4-fpm.sock"
	cmdGetFpmPath, err := exec.Command(
		"cat",
		"/etc/php/$(php -r 'echo PHP_VERSION;' | grep --only-matching --perl-regexp '7.\\d+')/fpm/pool.d/www.conf",
		"|",
		"grep",
		"'listen ='",
		).Output()
	// Replace php fpm socket path
	replaceContent(domainPath, "[phpFpmSocket]", "unix:/var/" + string(cmdGetFpmPath)[9:]);

	// Once successfully created into sites-available, create symlink to that file
	cmdLn := exec.Command("ln", "-s", domainPath, "/etc/nginx/sites-enabled/")
	cmdLn.Run()

	// Restart nginx web server once done
	cmdRestartWebServer := exec.Command("service", "nginx", "restart")
	cmdRestartWebServer.Run();
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
