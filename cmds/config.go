package main

import (
  "github.com/lyanchih/LazyKube"
  "github.com/spf13/cobra"
)

var (
  configFile string
  outputPath string
)

const configUsage = `
Generate deploy config
`

func newConfigCmd() *cobra.Command {
  cmd:= &cobra.Command{
    Use: "config",
    Short: "Generate deploy config",
    Long: configUsage,
    RunE: func(cmd *cobra.Command, args []string) error {
      c, err := lazy.Load(configFile);
      if err != nil {
        return err
      }
      
      return c.Generate(outputPath)
    },
  }

  f := cmd.Flags()
  f.StringVar(&configFile, "config-file", "etc/lazy.ini", "Lazykube ini config file")
  f.StringVar(&outputPath, "output", "_output", "Deploy config output path")
  
  return cmd
}
