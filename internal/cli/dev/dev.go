// Package dev hosts the developer toolkit subcommands:
// init wizard, static serve, script runner, git helpers.
package dev

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:     "dev",
		Aliases: []string{"d"},
		Short:   "Developer toolkit (init, serve, run, git, http, env, jwt)",
	}
	c.AddCommand(initCmd(), serveCmd(), runCmd(), gitCmd(), envCmd(), httpCmd(), jwtCmd())
	return c
}
