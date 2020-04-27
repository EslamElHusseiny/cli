package command

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cli/cli/utils"
)

func init() {
	RootCmd.AddCommand(creditsCmd)
}

var creditsCmd = &cobra.Command{
	Use:   "credits",
	Short: "TODO",
	Long:  "TODO",
	RunE:  credits,
}

func credits(cmd *cobra.Command, args []string) error {
	ctx := contextForCommand(cmd)
	client, err := apiClientForContext(ctx)
	if err != nil {
		return err
	}

	baseRepo, err := determineBaseRepo(cmd, ctx)
	if err != nil {
		return err
	}

	type Contributor struct {
		// TODO silly idea: render their avatar. maybe in a cute grid?
		// https://github.com/eliukblau/pixterm
		Login string
	}

	type Result []Contributor

	result := Result{}
	body := bytes.NewBufferString("")
	path := fmt.Sprintf("repos/%s/%s/contributors", baseRepo.RepoOwner(), baseRepo.RepoName())

	err = client.REST("GET", path, body, &result)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	isTTY := false
	if outFile, isFile := out.(*os.File); isFile {
		isTTY = utils.IsTerminal(outFile)
		if isTTY {
			// FIXME: duplicates colorableOut
			out = utils.NewColorable(outFile)
		}
	}

	logins := []string{}
	for _, c := range result {
		if !isTTY {
			fmt.Fprintf(out, "%s\n", c.Login)
		} else {
			logins = append(logins, c.Login)
		}
	}

	if !isTTY {
		return nil
	}

	// TODO pretty list

	return nil
}
