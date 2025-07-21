package commands

import (
	"fmt"

	"github.com/orchestre-dev/ccproxy/internal/utils"
	"github.com/spf13/cobra"
)

// EnvCmd returns the env command
func EnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Display all CCProxy environment variables",
		Long:  "Display documentation for all environment variables supported by CCProxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(utils.GetEnvironmentVariableDocumentation())
			return nil
		},
	}

	return cmd
}
