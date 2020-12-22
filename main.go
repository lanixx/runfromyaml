package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/ionrock/procs"

	"gopkg.in/yaml.v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func exeCommandWithinBash(cmd []string, desc string, wg *sync.WaitGroup) {
	color.New(color.FgGreen).Println("==> " + desc)
	fmt.Println(cmd)
	var bash []string
	bash = append([]string{"bash", "-c"}, strings.Join(cmd, " "))
	command := exec.Command(bash[0], bash[1:]...)
	command.Env = os.Environ()
	color.New(color.FgYellow).Println("Command:", bash, "\n")
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	fmt.Sprintln(bash)

	if err := command.Run(); err != nil {
		log.Fatalf("Start: %v", err)
	}
	wg.Done()
}

func exeCommand(cmd []string, desc string, wg *sync.WaitGroup) {
	color.New(color.FgGreen).Println("==> " + desc)
	command := exec.Command(cmd[0], cmd[1:]...)
	command.Env = os.Environ()
	color.New(color.FgYellow).Println("Command:", cmd, "\n")
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		log.Fatalf("Start: %v", err)
	}
	wg.Done()
}

func exeCommandTest(cmd []string, desc string, wg *sync.WaitGroup) {

	// define command set
	cmds := []*exec.Cmd{
		exec.Command(cmd[0], cmd[1:]...),
	}

	// init procs with command set
	p := procs.Process{Cmds: cmds}

	// parse environment variables
	env := procs.ParseEnv(os.Environ())
	p.Env = env

	// prepare output handler
	p.OutputHandler = func(line string) string {
		color.New(color.FgGreen).Println("==> " + desc)
		color.New(color.FgYellow).Println("Command: ", cmd)
		return line
	}

	// prepare error handler
	p.ErrHandler = func(line string) string {
		color.New(color.FgRed).Println("Command: ", cmd)
		return line
		fmt.Println(cmds)
		fmt.Println(p)
		fmt.Println(env)
		os.Exit(0)
		return line
	}

	color.New(color.FgGreen).Println("==> " + desc)
	color.New(color.FgYellow).Println("Command: ", cmd)

	p.Run()
	p.Wait()

	color.New(color.FgBlue).Println(cmds)
	color.New(color.FgBlue).Println(p)
	color.New(color.FgBlue).Println(env)

	out, _ := p.Output()
	fmt.Printf(string(out))
	wg.Done()
}

func writeFile(file string, path string, perm os.FileMode) {
	bytefile := []byte(file)
	err := ioutil.WriteFile(path, bytefile, perm)
	check(err)
}

func readFile(file string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("File contents: %s", content)
}

func remove(slice []string, i int) []string {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

func main() {
	var results map[interface{}][]interface{}
	var types map[interface{}]interface{}
	var values string
	var cmds []string
	var desc string
	var confdata string
	var confdest string
	var confperm os.FileMode
	var file string
	var help bool
	var debug bool

	programm := os.Args

	// parse flags
	flag.StringVar(&file, "file", "commands.yaml", "input config filename")
	flag.BoolVar(&help, "help", false, "Display this help")
	flag.BoolVar(&debug, "debug", false, "Debug Mode")
	flag.Parse()

	red := color.New(color.FgRed).PrintfFunc()
	yellow := color.New(color.FgYellow).PrintfFunc()
	green := color.New(color.FgGreen).PrintfFunc()
	blue := color.New(color.FgBlue).PrintfFunc()
	white := color.New(color.FgHiWhite).PrintfFunc()
	cyan := color.New(color.FgHiCyan).PrintfFunc()

	if debug {
		red("\n", programm)
	}

	yamlFile, err := ioutil.ReadFile(file)
	err = yaml.Unmarshal(yamlFile, &results)
	for key := range results["cmd"] {
		if !reflect.ValueOf(results["cmd"][key].(map[interface{}]interface{})).IsNil() {
			types = results["cmd"][key].(map[interface{}]interface{})
			if debug {
				cyan("\n\n%+v\n\n", types)
				blue("Name: %+v\n", types["name"])
				blue("Beschreibung: %+v\n", types["desc"])
				blue("Key: %+v\n", key)
				blue("Command: %+v\n", values)
				blue("Data:\n---\n%+v\n---\n", types["confdata"])
				blue("Destination: %+v\n", types["confdest"])
				blue("Permissions: %+v\n", types["confperm"])
				fmt.Printf("\n")
			}
			if types["type"] == "exec" {
				if debug {
					white("Key: %+v\n", key)
					green("Name: %+v\n", types["name"])
					green("Beschreibung: %+v\n", types["desc"])
					yellow("Command: %+v\n", values)
					fmt.Printf("\n")
				}
				wg := new(sync.WaitGroup)
				if !reflect.ValueOf(types["values"].(interface{})).IsNil() {
					values = fmt.Sprintf("%v", types["values"].(interface{}))
					values = strings.TrimPrefix(values, "[")
					values = strings.TrimSuffix(values, "]")
					cmds = strings.Fields(values)
				}
				if string(types["desc"].(string)) != "" {
					desc = fmt.Sprintf("%v", types["desc"])
				}
				wg.Add(1)
				go exeCommand(cmds, desc, wg)
				wg.Wait()
			}
			if types["type"] == "shell" {
				if debug {
					white("Key: %+v\n", key)
					green("Name: %+v\n", types["name"])
					green("Beschreibung: %+v\n", types["desc"])
					yellow("Command: %+v\n", values)
					fmt.Printf("\n")
				}
				wg := new(sync.WaitGroup)
				if !reflect.ValueOf(types["values"].(interface{})).IsNil() {
					values = fmt.Sprintf("%v", types["values"].(interface{}))
					values = strings.TrimPrefix(values, "[")
					values = strings.TrimSuffix(values, "]")
					cmds = strings.Fields(values)
				}
				if string(types["desc"].(string)) != "" {
					desc = fmt.Sprintf("%v", types["desc"])
				}
				wg.Add(1)
				go exeCommandWithinBash(cmds, desc, wg)
				wg.Wait()
			}
			if types["type"] == "conf" {
				if debug {
					yellow("Config: %+v\n", types["confdata"])
					yellow("Config: %+v\n", types["confdest"])
					yellow("Config: %+v\n", types["confperm"])
					fmt.Printf("\n")
				}
				if reflect.ValueOf(types["confdata"].(string)).String() != "" {
					confdata = types["confdata"].(string)
				}
				if reflect.ValueOf(types["confdest"].(string)).String() != "" {
					confdest = types["confdest"].(string)
				}
				if reflect.ValueOf(types["confperm"].(int)).Int() != 0 {
					confperm = os.FileMode(int(types["confperm"].(int)))
				}
				if confdata != "" && confdest != "" && string(confperm) != "" {
					writeFile(confdata, confdest, confperm)
					//readFile(string(confdest))
				}
			}
			fmt.Printf("\n")
		}
	}
	check(err)
}
