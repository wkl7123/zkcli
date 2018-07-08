package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/c-bata/go-prompt"
	"github.com/wkl7123/zkcli/core"
	"log"
)

var gitCommit = "unknown"
var built = "unknown"

const version = "0.2.0"

type NewLogger struct{
	debugLog log.Logger
}

func (nl NewLogger) Printf(format string, a ...interface{}) {
	nl.debugLog.Printf(format, a...)
}

func main() {
	servers := flag.String("s", "127.0.0.1:2181", "Servers")
	username := flag.String("u", "", "Username")
	password := flag.String("p", "", "Password")
	showVersion := flag.Bool("version", false, "Show version info")
	flag.Parse()
	args := flag.Args()
	//fmt.Println(args)

	if *showVersion {
		fmt.Printf("Version:\t%s\nGit commit:\t%s\nBuilt: %s\n",
			version, gitCommit, built)
		os.Exit(0)
	}

	config := core.NewConfig(strings.Split(*servers, ","))
	if *username != "" && *password != "" {
		auth := core.NewAuth(
			"digest", fmt.Sprintf("%s:%s", *username, *password),
		)
		config.Auth = auth
	}
	conn, err := config.Connect()
	logFile,err  := os.Create("/tmp/zkcli.log")
	defer logFile.Close()
	fileLog := log.New(logFile,"[Debug]",log.LstdFlags)
	newLogger := NewLogger{debugLog: *fileLog}
	conn.SetLogger(newLogger)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	name, options := core.ParseCmd(strings.Join(args, " "))
	cmd := core.NewCmd(name, options, conn, config)
	if len(args) > 0 {
		cmd.ExitWhenErr = true
		cmd.Run()
		return
	}

	p := prompt.New(
		core.GetExecutor(cmd),
		core.GetCompleter(cmd),
		prompt.OptionTitle("zkcli: A interactive Zookeeper client"),
		prompt.OptionPrefix(">>> "),
	)
	p.Run()
}
