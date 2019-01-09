package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log"
	"os"
	"os/user"
	"strings"
)

const (
	versionString = "kubectl-repl {{{VERSION}}}"
)

var (
	input     *bufio.Reader
	namespace string
	context   string
	verbose   bool
)

func prompt() (string, error) {
	color.New(color.Bold).Print("# ")

	if context != "" {
		color.New(color.FgBlack, color.Italic).Print(context)
		fmt.Print(" ")
	}

	if namespace != "" {
		color.New(color.Bold).Print(namespace)
	} else {
		color.New(color.Bold).Print("namespace")
	}
	fmt.Print(" ")

	line, err := input.ReadString('\n')
	if err != nil {
		return "", err
	}
	response := strings.Trim(line, "\n")
	return substituteForVars(substituteForAliases(response))
}

func printIndexedLine(index, line string) {
	coloredIndex := color.New(color.FgBlue).Sprintf("$%s", index)
	fmt.Printf("%s \t%s\n", coloredIndex, line)
}

func repl(commands Commands) error {
	command, err := prompt()
	if err != nil {
		return err
	}

	for _, builtin := range commands {
		if builtin.filter(command) {
			return builtin.run(command)
		}
	}

	return sh(kubectl(command))
}

func main() {
	var version bool
	flag.BoolVar(&verbose, "verbose", false, "Verbose")
	flag.BoolVar(&version, "version", false, "Print current version")
	flag.StringVar(&context, "context", "", "Override current context")
	flag.Parse()

	if version {
		fmt.Println(versionString)
		return
	}

	usr, err := user.Current()
    if err != nil {
        log.Fatal( err )
    }
	err = loadAliasesFromFile(strings.Join([]string{usr.HomeDir, ".kubectlrepl"}, "/"))
	if err != nil {
		log.Fatal(err)
	}

	commands := Commands{
		builtinExit{},
		builtinNamespace{},
		builtinShell{},
		builtinGet{},
	}
	err = commands.Init()
	if err != nil {
		log.Fatal(err)
	}

	variables = make(map[string][]string)
	input = bufio.NewReader(os.Stdin)

	err = pickNamespace()
	if err == io.EOF {
		return
	} else if err != nil {
		log.Fatal(err)
	}

	for {
		err = repl(commands)
		if err == io.EOF {
			break
		} else if err != nil {
			color.New(color.FgRed).Println(err)
		}
	}
}
