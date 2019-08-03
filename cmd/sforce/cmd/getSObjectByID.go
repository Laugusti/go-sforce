package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var fields string

// getSObjectByIDCmd represents the getSObjectByID command
var getSObjectByIDCmd = &cobra.Command{
	Use:   "getSObjectByID <name> <id>",
	Args:  cobra.ExactArgs(2),
	Short: "Retrieves the SObject using the Object Name and Object ID",
	Run: func(cmd *cobra.Command, args []string) {
		sobj, err := restClient.GetSObject(args[0], args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// create json encoder to write SObject to stdout
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		if err := enc.Encode(sobj); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	restCmd.AddCommand(getSObjectByIDCmd)
	// TODO: add fields query to client
	getSObjectByIDCmd.Flags().StringVarP(&fields, "fields", "f", "", "Specify the fields you want to retrieve")
}
