package main

import (
  "fmt"
  "os"

  flag "github.com/spf13/pflag"
  "github.com/malcolmholmes/grafana-dash/pkg/dash"
)

func usage() {
  fmt.Println("Usage: g <cmd> <file>")
  os.Exit(1)
}

func main() {
  flag.Parse()
  args := flag.Args()
  if len(args) != 2 {
    usage()
  }
  cmd:=args[0]
  jsonnetFile := args[1]

  config, err := dash.ParseEnvironment()
  if err != nil {
    fmt.Println("ERROR", err)
    os.Exit(1)
  }
  
  if (cmd == "get") {
    err = dash.Get(*config, jsonnetFile)
  } else if (cmd == "show") {
    err = dash.Show(*config, jsonnetFile)
  } else if (cmd == "diff") {
    err = dash.Diff(*config, jsonnetFile)
  } else if (cmd == "apply") {
    err = dash.Apply(*config, jsonnetFile)
  } else {
    fmt.Printf("Unknown command: %s\n", cmd)
    os.Exit(1)
  }
  if err != nil {
    fmt.Println("ERROR", err)
    os.Exit(1)
  }
}