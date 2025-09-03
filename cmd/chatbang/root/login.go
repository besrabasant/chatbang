package root

import (
    "gg/pkg/app"
    "github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
    Use:   "login",
    Short: "Open ChatGPT to set up/login and grant clipboard permission",
    RunE: func(cmd *cobra.Command, args []string) error {
        a := app.New()
        return a.Login()
    },
}

func init() {
    rootCmd.AddCommand(loginCmd)
}

