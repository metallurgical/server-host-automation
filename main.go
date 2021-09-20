package main

import (
	"fmt"
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
	var domainPath = "/etc/nginx/site-available/" + domain + ".conf"

	cmd := exec.Command("ln", "-s", domainPath, "/etc/nginx/sites-enabled/")
	cmd.Run()

	cmd = exec.Command("ln", "-s", domainPath, "/etc/nginx/sites-enabled/")
	cmd.Run()
}
