package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
		from, err := os.Open(filepath.Join(from, file.Name()))
		if err != nil {
			log.Fatal(err)
			from.Close()
			return err
		}

		to, err := os.OpenFile(filepath.Join(to, dirToSaveIn, file.Name()), os.O_RDWR|os.O_CREATE, file.Mode())
		if err != nil {
			log.Fatal(err)
			from.Close()
			to.Close()
			return err
		}

		_, err = io.Copy(to, from)
		if err != nil {
			from.Close()
			to.Close()
			log.Fatal(err)
			return err
		}
		from.Close()
		to.Close()
		log.Println(color.GreenString("done"))
	}
	return nil
}
