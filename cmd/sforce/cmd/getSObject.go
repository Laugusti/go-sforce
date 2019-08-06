package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

var fields string

// getSObjectCmd represents the getSObject command
var getSObjectCmd = &cobra.Command{
	Use:   "getSObject <name> <id>",
	Args:  cobra.ExactArgs(2),
	Short: "Retrieves the SObject using the Object Name and Object ID",
	Run: func(cmd *cobra.Command, args []string) {
		input := &restapi.GetSObjectInput{
			SObjectName: args[0],
			SObjectID:   args[1],
			Fields:      strings.Split(fields, ","),
		}
		out, err := restClient.GetSObject(input)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// create json encoder to write SObject to stdout
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		if err := enc.Encode(out.SObject); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	restCmd.AddCommand(getSObjectCmd)
	getSObjectCmd.Flags().StringVarP(&fields, "fields", "f", "", "Specify the fields you want to retrieve")
}
