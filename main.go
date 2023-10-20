package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
)

var configFile string
var testQuery bool
var checkConfig bool

func main() {
	cli := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	cli.StringVarP(&configFile, "config", "c", "", "Read configuration from this yaml file")
	cli.BoolVar(&checkConfig, "check", false, "Check configuration file syntax")
	cli.BoolVar(&testQuery, "test", false, "Test a single query instead of running as a server")

	err := cli.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}
	if configFile == "" {
		_, _ = fmt.Fprintf(os.Stderr, "A configuration file must be provided with --config=\n")
		os.Exit(2)
	}
	config, err := LoadConfig(configFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to load configuration: %s\n", err)
		os.Exit(2)
	}
	if checkConfig {
		os.Exit(0)
	}
	if testQuery {
		if len(cli.Args()) != 2 {
			_, _ = fmt.Fprintf(os.Stderr, "Expected parameters <hostname> <type>")
			os.Exit(2)
		}
		qname := strings.TrimSuffix(cli.Arg(0), ".")
		qtype := strings.ToUpper(cli.Arg(1))
		responses := Respond(config, qname, qtype, "1")
		for _, response := range responses {
			fmt.Printf("  %s\n", response.PipeResponse("1"))
		}
		os.Exit(0)
	}
	err = ServePipe(config, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exiting due to error: %s\n", err)
	}
}
