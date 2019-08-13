package cmd

import (
	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

// upsertSObjectByExternalIDCmd represents the upsertByExternalId command
var upsertSObjectByExternalIDCmd = &cobra.Command{
	Use:   "upsertByExternalId <name> <extidfield> <extid> [<file>]",
	Args:  cobra.RangeArgs(3, 4),
	Short: "Create/Update an existing SObject using the Object Name, External ID Field, External ID and data file",
	Long: `Create/Update an existing SObject using the Object Name, External ID Field, External ID and data file.
With no file or when file is -, read standard input`,
	Run: func(cmd *cobra.Command, args []string) {
		// get file argument or use -
		file := "-"
		if len(args) == 4 {
			file = args[3]
		}

		// unmarshal file to sobject
		var sobj restapi.SObject
		unmarshalJSONFile("UpsertSObjectByExternalID", file, &sobj)

		// create api input
		input := &restapi.UpsertSObjectByExternalIDInput{
			SObjectName:     args[0],
			ExternalIDField: args[1],
			ExternalID:      args[2],
			SObject:         sobj,
		}

		// do api request
		out, err := restClient.UpsertSObjectByExternalID(input)
		exitIfError("UpsertSObjectByExternalIDInput", err)

		// write result to stdout
		marshalJSONToStdout("UpsertSObjectByExternalIDInput", out.Result)
	},
}

func init() {
	sobjectCmd.AddCommand(upsertSObjectByExternalIDCmd)
}
