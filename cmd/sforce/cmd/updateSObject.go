package cmd

import (
	"fmt"

	restapi "github.com/Laugusti/go-sforce/api/rest"
	"github.com/spf13/cobra"
)

// updateSObjectCmd represents the updateSObject command
var updateSObjectCmd = &cobra.Command{
	Use:   "updateSObject <name> <id> [<file>]",
	Args:  cobra.RangeArgs(2, 3),
	Short: "Updates an existing SObject using the Object Name, Object ID and data file",
	Long: `Updates an existing SObject using the Object Name, Object ID and data file.
With no file or when file is -, read standard input.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get file argument or use -
		file := "-"
		if len(args) == 3 {
			file = args[2]
		}

		// unmarshal file to sobject
		var sobj restapi.SObject
		unmarshalJSONFile("UpdateSObject", file, &sobj)

		// create input
		input := &restapi.UpdateSObjectInput{
			SObjectName: args[0],
			SObjectID:   args[1],
			SObject:     sobj,
		}

		// do api request
		_, err := restClient.UpdateSObject(input)
		exitIfError("UpdateSObject", err)

		// print success message
		fmt.Printf("Updated %s object with Id %q\n", args[0], args[1])
	},
}

func init() {
	sobjectCmd.AddCommand(updateSObjectCmd)
}
