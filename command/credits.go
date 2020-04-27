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

	// TODO rainbow loop text

	for x, l := range logins {
		// TODO something cute pre/post
		fmt.Fprintf(out, "%s\n", getColor(x)(l))
	}

	// TODO pretty list

	return nil
}

func getColor(x int) func(string) string {
	rainbow := []func(string) string{
		utils.Magenta,
		utils.Red,
		utils.Yellow,
		utils.Green,
		utils.Cyan,
		utils.Blue,
	}

	ix := x % len(rainbow)

	return rainbow[ix]

}
