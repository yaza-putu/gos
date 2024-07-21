// Author : Yaza
//
// GOS -> Golang Starter API Tool
// Gos is a tool used to create new projects from project templates that we have provided
//
// echo template : https://github.com/yaza-putu/golang-starter-mongo-api
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type projectInfo struct {
	Modul       string
	ProjectName string
}

const (
	ECHO_GORM  = "https://github.com/yaza-putu/golang-starter-api.git"
	ECHO_MONGO = "https://github.com/yaza-putu/golang-starter-mongo-api.git"
)

const LOGO = `
   ______      __                     _____ __             __               ___    ____  ____
  / ____/___  / /___ _____  ____ _   / ___// /_____ ______/ /____  _____   /   |  / __ \/  _/
 / / __/ __ \/ / __ / __ \/ __ /   \__ \/ __/ __ / ___/ __/ _ \/ ___/  / /| | / /_/ // /  
/ /_/ / /_/ / / /_/ / / / / /_/ /   ___/ / /_/ /_/ / /  / /_/  __/ /     / ___ |/ ____// /   
\____/\____/_/\__,_/_/ /_/\__, /   /____/\__/\__,_/_/   \__/\___/_/     /_/  |_/_/   /___/   
                           /____/                                                               
`

func main() {
	command := flag.NewFlagSet("create", flag.ExitOnError)
	echoFlag := command.Bool("echo", false, "Create new project using the echo gorm golang starter api")
	echoMongoFlag := command.Bool("mongo", false, "Create new project using the echo mongo golang starter api")
	flag.Parse()

	fmt.Println(LOGO)

	p := projectInfo{}

	if len(os.Args) == 1 {
		fmt.Println("suggestion \n commands:  \n - create \n\n flag : \n - echo \n - mongo")
		fmt.Println("ex:gos create --echo")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "create":
		p.ProjectName = getInput("Project Name")
		p.Modul = getInput("Module Name")

		// parse flag
		command.Parse(os.Args[2:])

		repoUrl := ""

		if *echoFlag && *echoMongoFlag {
			repoUrl = ECHO_MONGO
		} else if *echoFlag {
			repoUrl = ECHO_GORM
		}

		if repoUrl == "" {
			fmt.Println("Unknown flag.")
			os.Exit(1)
		}

		loadingDone := make(chan bool)
		go showLoading(loadingDone)

		cloneRepo(repoUrl)
		repo, err := extractRepoName(repoUrl)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = setup(p, repo)
		if err != nil {
			fmt.Printf("Failed create the project : %v", err)
			os.RemoveAll(repo)
		} else {
			loadingDone <- true
		}
	default:
		fmt.Println("Unknown command.")
		os.Exit(1)
	}
}

func setup(info projectInfo, folder string) error {
	// rename folder
	err := os.Rename(folder, info.ProjectName)
	if err != nil {
		return err
	}

	// change active directory
	err = os.Chdir(info.ProjectName)
	if err != nil {
		return err
	}

	// remove git
	err = runCommand("rm", "-rf", ".git")
	if err != nil {
		return err
	}

	// go mod tidy
	err = runCommand("go", "mod", "tidy")
	if err != nil {
		return err
	}

	// setup env
	err = runCommand("cp", ".env.example", ".env")
	if err != nil {
		return err
	}

	// setup env
	err = runCommand("cp", ".env.example", ".env.test")
	if err != nil {
		return err
	}

	// setup module
	err = runCommand("go", "run", "cmd/zoro.go", "key:generate")
	if err != nil {
		return err
	}

	// setup module
	err = runCommand("go", "run", "cmd/zoro.go", "configure:module", info.Modul)
	if err != nil {
		return err
	}

	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func showLoading(done <-chan bool) {
	fmt.Print("Creating project")
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Print(".")
		case <-done:
			fmt.Println("\nDone.")
			return
		}
	}
}

func cloneRepo(repoURL string) error {
	repoName, err := extractRepoName(repoURL)
	if err != nil {
		return fmt.Errorf("failed to extract repository name: %v", err)
	}

	if folderExists(repoName) {
		fmt.Printf("Folder '%s' already exists. Removing...\n", repoName)
		err := os.RemoveAll(repoName)
		if err != nil {
			return fmt.Errorf("failed to remove existing folder: %v", err)
		}
	}

	cmd := exec.Command("git", "clone", repoURL)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error cloning repository:", err)
	}
	return err
}

func folderExists(folder string) bool {
	info, err := os.Stat(folder)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func extractRepoName(repoURL string) (string, error) {
	base := filepath.Base(repoURL)
	if len(base) == 0 {
		return "", fmt.Errorf("invalid repository URL")
	}
	return base[:len(base)-len(filepath.Ext(base))], nil
}

func getInput(prompt string) string {
	fmt.Print(prompt + ": ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}
	input = strings.TrimSpace(input)
	return input
}
