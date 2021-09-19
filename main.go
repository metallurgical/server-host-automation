package main

import "fmt"

var domain, projectRoot, gitEndpoint string;

func main() {
	askForInput();
}

func askForInput() {
	fmt.Println("[ Server host automation tools by @metallurgical(https://github.com/metallurgical) ]");
	fmt.Print("1) What is the domain name (without http/https) ? : ");
	fmt.Scanln(&domain);

	fmt.Print("2) Where to clone the git repo (Full path to base folder of project) ? : ");
	fmt.Scanln(&projectRoot);

	fmt.Print("3) Full URL of git endpoint (will be automatically cloned into folder set in no 2) ? : ");
	fmt.Scanln(&gitEndpoint);
}
