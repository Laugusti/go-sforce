package cmd

import (
	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

var sobjExtIDFields string

// getSObjectByExternalIDCmd represents the getSObjectByExternalID command
var getSObjectByExternalIDCmd = &cobra.Command{
	Use:   "getSObjectByExternalID <name> <field> <id>",
	Args:  cobra.ExactArgs(3),
	Short: "Retrieves the SObject using the Object Name, External ID Field and External ID",
	Run: func(cmd *cobra.Command, args []string) {
		// create api input
		input := &restapi.GetSObjectByExternalIDInput{
			SObjectName:     args[0],
			ExternalIDField: args[1],
			ExternalID:      args[2],
			Fields:          splitString(sobjExtIDFields, ","),
		}

		// do api request
		out, err := restClient.GetSObjectByExternalID(input)
		exitIfError("GetSObjectByExternalID", err)

		// write sobject to stdout
		marshalJSONToStdout("GetSObjectByExternalID", out.SObject)
	},
}

func init() {
	sobjectCmd.AddCommand(getSObjectByExternalIDCmd)
	getSObjectByExternalIDCmd.Flags().StringVarP(&sobjExtIDFields, "fields", "f", "", "Specify the fields you want to retrieve")
}
