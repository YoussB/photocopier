package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	flags "github.com/jessevdk/go-flags"
)

func main() {
	cmd := &CopierCommand{}

	parser := flags.NewParser(cmd, flags.Default)
	parser.Command.Find("copy")
	_, err := parser.Parse()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

type CopierCommand struct {
	CopyCommand CopyCommand `command:"copy"`
}

type CopyCommand struct {
	InputDir  string `short:"i" long:"input-directory" default:"." description:"The directory from which I would read the photos"`
	OutputDir string `short:"o" long:"output-directory" default:"." description:"The directory from which I would put the photos' folder in"`
}

func (c *CopyCommand) Execute(args []string) error {
	return copyFiles(c.InputDir, c.OutputDir)
}

func copyFiles(from, to string) error {
	dirsInOutput := make(map[string]bool)

	files, err := ioutil.ReadDir(from)
	if err != nil {
		return fmt.Errorf("input read error: %w", err)
	}

	if _, err := os.Stat(to); os.IsNotExist(err) {
		if err := os.MkdirAll(to, 0755); err != nil {
			return err
		}
	}
	outputDirs, err := ioutil.ReadDir(to)
	if err != nil {
		return fmt.Errorf("output read error: %w", err)
	}
	for _, dir := range outputDirs {
		if dir.IsDir() {
			dirsInOutput[dir.Name()] = true
		}
	}

	for _, file := range files {
		if file.IsDir() {
			if err := copyFiles(filepath.Join(from, file.Name()), to); err != nil {
				return err
			}
			continue
		}
		log.Println(filepath.Join(from, file.Name()), "👮‍♂️ying")
		dirToSaveIn := file.ModTime().Format("20060102(Mon Jan 02 2006)")
		if !dirsInOutput[dirToSaveIn] {
			os.MkdirAll(filepath.Join(to, dirToSaveIn), 0755)
			dirsInOutput[dirToSaveIn] = true
		}
		if _, err := os.Stat(filepath.Join(to, dirToSaveIn, file.Name())); err == nil {
			log.Println(color.RedString("^^ file exists.. skipping"))
			continue
		}
		if err := copyFile(filepath.Join(from, file.Name()), filepath.Join(to, dirToSaveIn, file.Name()), file.Mode()); err != nil {
			return err
		}
		log.Println(color.GreenString("done"))
	}
	return nil
}

func copyFile(from, to string, mode os.FileMode) error {
	if runtime.GOOS == `linux` {
		cpCmd := exec.Command("cp", "-rf", from, to)
		return cpCmd.Run()
	} else if runtime.GOOS == `darwin` {
		cpCmd := exec.Command("ditto", from, to)
		return cpCmd.Run()
	} else {
		fromFile, err := os.Open(from)
		if err != nil {
			log.Fatal(err)
			return err
		}
		defer fromFile.Close()

		toFile, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, mode)
		if err != nil {
			log.Fatal(err)
			return err
		}
		defer toFile.Close()

		_, err = io.Copy(toFile, fromFile)
		if err != nil {
			log.Fatal(err)
			return err
		}
		return nil
	}
}
