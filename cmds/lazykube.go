package main

import (
  "os"
  "github.com/spf13/cobra"
)

const globalUsage = `The Lazy deploy tool for kuberentes cluster
Common actions from this point include:
- lazykube config:      Generate deploy config
`

func newRootCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use: "lazykube",
    Short: "The Lazy deploy tool for kuberentes cluster",
    Long: globalUsage,
  }

  cmd.AddCommand(newConfigCmd())
  
  return cmd
}

func main() {
  cmd := newRootCmd()
  
  if err := cmd.Execute(); err != nil {
    os.Exit(1)
  }
}
