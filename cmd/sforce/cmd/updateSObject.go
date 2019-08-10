package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

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
		// get reader (stdin or file)
		var r io.Reader = os.Stdin
		if len(args) == 3 && args[2] != "-" {
			b, err := ioutil.ReadFile(args[2])
			//if err == os.ErrNotExist{}
			exitIfError("UpdateSObject", err)
			r = bytes.NewReader(b)
		}
		// unmarshal to SObject
		var sobj restapi.SObject
		exitIfError("UpdateSObject", json.NewDecoder(r).Decode(&sobj))

		// do api request
		input := &restapi.UpdateSObjectInput{
			SObjectName: args[0],
			SObjectID:   args[1],
			SObject:     sobj,
		}
		_, err := restClient.UpdateSObject(input)
		exitIfError("UpdateSObject", err)

		fmt.Printf("Updated %s object with Id %q\n", args[0], args[1])
	},
}

func init() {
	restCmd.AddCommand(updateSObjectCmd)
}
