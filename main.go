package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const NginxSitesAvailablePath = "/etc/nginx/sites-available"
const ApacheSitesAvailablePath = "/etc/apache2/sites-available"

var (
	domain         string
	projectRoot    string
	gitEndpoint    string
	whichWebServer string
)

func main() {
	fmt.Println("[Server host automation tool by @metallurgical(https://github.com/metallurgical)]")
	askForInput()

	if projectRoot != "" && gitEndpoint != "" {
		cloneGitRepo()
	}

	switch whichWebServer {
	case "2":
		createNginxVhost()
		break
	case "1":
		createApacheVhost()
		break
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

	// Not really need this by now can get it directly from `php -v` command
	// fmt.Print("5) PHP version currently use? (Put only first two major number, eg: 7.4, 8.0, 7.2, 5.6) ? : ")
	// fmt.Scanln(&phpVersion)
}

func cloneGitRepo() {
	fmt.Println(">>>> Cloning git repository into " + projectRoot)
	if isExist, _ := exists(projectRoot); isExist == true {
		fmt.Println(">>>> Project directory path already exist. Skip")
	} else {
		cmdCloneRepo := exec.Command("git", "clone", gitEndpoint, projectRoot)
		cmdCloneRepo.Run()
		fmt.Println(">>>> Done clone git repository")
	}

	// Change ownership of storage folder
	fmt.Println(">>>> Change owner storage folder of " + projectRoot + "/storage as www-data/apache user")
	webServerUser,_ := getWebServerUser();
	cmdChangeUser := exec.Command("chown", webServerUser + ":" + webServerUser, "-R", projectRoot+"/storage")
	cmdChangeUser.Run()
	fmt.Println(">>>> Done change ownership of storage folder")

	// Copy .env.example file
	fmt.Println(">>>> Copy .env.example's content into .env file")
	if isExist, _ := exists(projectRoot + "/.env"); isExist == true {
		fmt.Println(">>>> Env file already exists. Skip")
	} else {
		cmdCopyEnv := exec.Command("cp", projectRoot+"/.env.example", projectRoot+"/.env")
		cmdCopyEnv.Run()
		fmt.Println(">>>> Done copied")
	}

	// Change directory easier to run any command related to project
	os.Chdir(projectRoot)

	// Run composer install
	fmt.Println(">>>> Running composer install, this might take a while.")
	if isExist, _ := exists(projectRoot + "/vendor"); isExist == true {
		fmt.Println(">>>> Vendor folder already exist. Skip.")
	} else {
		cmdComposer := exec.Command("composer", "install")
		cmdComposer.Run()
		fmt.Println(">>>> Done install composer dependencies")
	}

	// Run php generate key
	fmt.Println(">>>> Generate new APP_KEY")
	cmdPhpKey := exec.Command("php", "artisan", "key:generate")
	cmdPhpKey.Run()
	fmt.Println(">>>> Done generating APP_KEY")
}

// createNginxVhost will create a minimal server block for apache
// to run the desire domain for laravel project.
func createNginxVhost() {
	fmt.Println(">>>> Check if " + NginxSitesAvailablePath + " folder is exist..")
	if _, err := os.Stat(NginxSitesAvailablePath); err != nil {
		fmt.Println(">>>> Folder " + NginxSitesAvailablePath + " does not exist!. Abort")
		os.Exit(1)
		return
	}

	fmt.Println(">>>> Create nginx server block for domain: " + domain + ".")
	// Get/download the file from source
	cmdWget := exec.Command("wget", "https://raw.githubusercontent.com/metallurgical/server-host-automation/master/default-nginx-host.conf", "-P", "/tmp")
	cmdWget.Run()

	// Move the file from source into nginx sites-available folder
	var vhostFileName = domain + ".conf"
	var domainPath = NginxSitesAvailablePath + "/" + vhostFileName
	if isExist, _ := exists(domainPath); isExist == true {
		fmt.Println(">>>> Server block already exist. Skip")
	} else {
		cmdCp := exec.Command("mv", "/tmp/default-nginx-host.conf", domainPath)
		cmdCp.Run()

		// Replace document root full path into new project directory path
		replaceContent(domainPath, "[documentRoot]", projectRoot+"/public")
		// Replace matching server name with new the exact domain name
		replaceContent(domainPath, "[serverName]", domain)
		// Get the full path of php-fpm socket. This will return the output
		// something similar to this eg: "listen = /run/php/php7.4-fpm.sock"
		// cmdGetFpmPath, err := exec.Command("bash", "-c", "cat /etc/php/"+phpVersion+"/fpm/pool.d/www.conf | grep 'listen ='").Output()
		version, _ := getPhpVersion()
		cmdGetFpmPath, err := exec.Command("bash", "-c", "cat /etc/php/"+version+"/fpm/pool.d/www.conf | grep 'listen ='").Output()
		if err != nil {
			return
		}

		// Replace php fpm socket path
		replaceContent(domainPath, "[phpFpmSocket]", "unix:/var"+strings.TrimSpace(string(cmdGetFpmPath)[9:]))
		fmt.Println(">>>> Done creating server block")
		// Once successfully created into sites-available, create symlink to that file
		fmt.Println(">>>> Create symlink server block for domain: " + domain)
		cmdLn := exec.Command("ln", "-s", domainPath, "/etc/nginx/sites-enabled/")
		cmdLn.Run()

		// Restart nginx web server once done
		fmt.Println(">>>> Reloading web server to take effect of new changes")
		cmdRestartWebServer := exec.Command("service", "nginx", "reload")
		cmdRestartWebServer.Run()
		fmt.Println(">>>> Project are available to browse with new domain: " + domain)
	}
}

// createApacheVhost will create a minimal server block for apache
// to run the desire domain for laravel project.
func createApacheVhost() {
	fmt.Println(">>>> Check if " + ApacheSitesAvailablePath + " folder is exist..")
	if _, err := os.Stat(ApacheSitesAvailablePath); err != nil {
		fmt.Println(">>>> Folder " + ApacheSitesAvailablePath + " does not exist!. Abort")
		os.Exit(1)
		return
	}

	fmt.Println(">>>> Create apache server block for domain: " + domain + ".")
	// Get/download the file from source
	cmdWget := exec.Command("wget", "https://raw.githubusercontent.com/metallurgical/server-host-automation/master/default-apache-host.conf", "-P", "/tmp")
	cmdWget.Run()

	// Move the file from source into apache2 sites-available folder
	var vhostFileName = domain + ".conf"
	var domainPath = ApacheSitesAvailablePath + "/" + vhostFileName
	if isExist, _ := exists(domainPath); isExist == true {
		fmt.Println(">>>> Server block already exist. Skip")
	} else {
		cmdCp := exec.Command("mv", "/tmp/default-apache-host.conf", domainPath)
		cmdCp.Run()

		// Replace document root full path into new project directory path
		replaceContent(domainPath, "[documentRoot]", projectRoot+"/public")
		// Replace matching server name with new the exact domain name
		replaceContent(domainPath, "[serverName]", domain)

		fmt.Println(">>>> Done creating server block")
		// Once successfully created into sites-available, create symlink to that file
		fmt.Println(">>>> Create symlink server block for domain: " + domain)
		cmdLn := exec.Command("ln", "-s", domainPath, "/etc/apache2/sites-enabled/")
		cmdLn.Run()
		// Enable newly created site
		enSite := exec.Command("a2ensite", vhostFileName)
		enSite.Run()
		// Restart apache web server once done
		fmt.Println(">>>> Reloading web server to take effect of new changes")
		cmdRestartWebServer := exec.Command("service", "apache2", "restart")
		cmdRestartWebServer.Run()
		fmt.Println(">>>> Project are available to browse with new domain: " + domain)
	}
}

// replaceContent search for string and replace with the new one.
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

func executeCommand(command string) {
	cmdArgs := strings.Fields(command)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	oneByte := make([]byte, 100)
	num := 1
	for {
		if _, err := stdout.Read(oneByte); err != nil {
			fmt.Printf(err.Error())
			break
		}
		r := bufio.NewReader(stdout)
		line, _, _ := r.ReadLine()
		fmt.Println(string(line))
		num = num + 1
		if num > 3 {
			os.Exit(0)
		}
	}
	cmd.Wait()
}

// exists returns whether the given file or directory exists
// credit to https://stackoverflow.com/a/10510783
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// getPhpVersion retrieve automatically php version and extract only
// first 2 major version
func getPhpVersion() (string, error) {
	cmdPhpVersion, err := exec.Command("php", "-v").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(cmdPhpVersion)[4:7]), nil
}

// getWebServerUser retrieve web server's user that running as
// not using `ps aux` since its returning a lot of text
func getWebServerUser() (string, error) {
	var webServer = "nginx"
	if whichWebServer != "2" {
		webServer = "httpd|apache2|apache"
	}
	cmdUser, err := exec.Command("bash", "-c", "ps -ef | egrep '("+webServer+")' | grep -v `whoami` | grep -v root | head -n1 | awk '{print $1}'").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(cmdUser)), nil
}
