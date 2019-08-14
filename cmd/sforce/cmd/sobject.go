package cmd

import (
	"github.com/spf13/cobra"
)

// sobjectCmd represents the sobject command
var sobjectCmd = &cobra.Command{
	Use:   "sobject",
	Short: "The sobject command performs CRUD operations for Salesforce Objects",
}

func init() {
	restCmd.AddCommand(sobjectCmd)
}
