package cmd

import (
	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

var sobjFields string

// getSObjectCmd represents the getSObject command
var getSObjectCmd = &cobra.Command{
	Use:   "getSObject <name> <id>",
	Args:  cobra.ExactArgs(2),
	Short: "Retrieves the SObject using the Object Name and Object ID",
	Run: func(cmd *cobra.Command, args []string) {
		// create api input
		input := &restapi.GetSObjectInput{
			SObjectName: args[0],
			SObjectID:   args[1],
			Fields:      splitString(sobjFields, ","),
		}

		// do api request
		out, err := restClient.GetSObject(input)
		exitIfError("GetSObject", err)

		// write sobject to stdout
		marshalJSONToStdout("GetSObject", out.SObject)
	},
}

func init() {
	sobjectCmd.AddCommand(getSObjectCmd)
	getSObjectCmd.Flags().StringVarP(&sobjFields, "fields", "f", "", "Specify the fields you want to retrieve")
}
