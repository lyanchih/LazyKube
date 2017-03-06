package main

import (
	"flag"
	"github.com/lyanchih/LazyKube"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config-file", "etc/lazy.ini", "Config file")
}

func main() {
	flag.Parse()

	c, err := lazy.Load(configFile);
  if err != nil {
    return
  }

  c.Generate()
}
