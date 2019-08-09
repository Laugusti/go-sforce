package cmd

import (
	"encoding/json"
	"os"

	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

var sobjExtIDFields string

// getSObjectByExternalIDCmd represents the getSObjectByExternalID command
var getSObjectByExternalIDCmd = &cobra.Command{
	Use:   "getSObjectByExternalID <name> <field> <id>",
	Args:  cobra.ExactArgs(3),
	Short: "Retrieves the SObject using the Object Name, External ID field and External ID",
	Run: func(cmd *cobra.Command, args []string) {
		input := &restapi.GetSObjectByExternalIDInput{
			SObjectName:     args[0],
			ExternalIDField: args[1],
			ExternalID:      args[2],
			Fields:          splitString(sobjExtIDFields, ","),
		}
		out, err := restClient.GetSObjectByExternalID(input)
		exitIfError("GetSObjectByExternalID", err)

		// create json encoder to write SObject to stdout
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "\t")
		exitIfError("GetSObjectByExternalID", enc.Encode(out.SObject))
	},
}

func init() {
	restCmd.AddCommand(getSObjectByExternalIDCmd)
	getSObjectByExternalIDCmd.Flags().StringVarP(&sobjExtIDFields, "fields", "f", "", "Specify the fields you want to retrieve")
}
