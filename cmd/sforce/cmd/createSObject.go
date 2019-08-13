package cmd

import (
	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

// createSObjectCmd represents the create command
var createSObjectCmd = &cobra.Command{
	Use:   "create <name> [<file>]",
	Args:  cobra.RangeArgs(1, 2),
	Short: "Creates a new SObject using Object Name and data file",
	Long: `Creates a new SObject using Object Name and data file.
With no file or when file is -, read standard input.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get file argument or use -
		file := "-"
		if len(args) == 2 {
			file = args[1]
		}

		// unmarshal file to sobject
		var sobj restapi.SObject
		unmarshalJSONFile("CreateSObject", file, &sobj)

		// create api input
		input := &restapi.CreateSObjectInput{
			SObjectName: args[0],
			SObject:     sobj,
		}

		// do api request
		out, err := restClient.CreateSObject(input)
		exitIfError("CreateSObject", err)

		// write result to stdout
		marshalJSONToStdout("CreateSObject", out.Result)
	},
}

func init() {
	sobjectCmd.AddCommand(createSObjectCmd)
}
