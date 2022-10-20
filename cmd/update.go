package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update bolter to the latest version",
	Long:  `Update bolter to the latest version.`,
	Run: func(cmd *cobra.Command, args []string) {
		updateCommand := exec.Command("bash", "-c", "curl -sfL https://raw.githubusercontent.com/hasit/bolter/main/install.sh | sh")
		updateCommand.Stdin = os.Stdin
		updateCommand.Stdout = os.Stdout
		updateCommand.Stderr = os.Stderr

		err := updateCommand.Run()
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	},
}
