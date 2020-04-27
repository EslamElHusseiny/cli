package command

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/cli/cli/utils"
)

// clear stuff taken from
// https://stackoverflow.com/questions/22891644/how-can-i-clear-the-terminal-screen-in-go
var clears map[string]func() //create a map for storing clear funcs

var thankYou = `
     _                    _
    | |                  | |
_|_ | |     __,   _  _   | |           __
 |  |/ \   /  |  / |/ |  |/_)   |   | /  \_|   |
 |_/|   |_/\_/|_/  |  |_/| \_/   \_/|/\__/  \_/|_/
                                   /|
                                   \|
                              _
                           o | |                           |
 __   __   _  _  _|_  ,_     | |        _|_  __   ,_    ,  |
/    /  \_/ |/ |  |  /  |  | |/ \_|   |  |  /  \_/  |  / \_|
\___/\__/   |  |_/|_/   |_/|_/\_/  \_/|_/|_/\__/    |_/ \/ o
                                                            
                                                            
`

func init() {

	clears = make(map[string]func()) //Initialize it
	clears["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clears["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

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
	outFile, isFile := out.(*os.File)
	if isFile {
		isTTY = utils.IsTerminal(outFile)
		if isTTY {
			// FIXME: duplicates colorableOut
			out = utils.NewColorable(outFile)
		}
	}

	// TODO probably can't animate on windows
	logins := []string{}
	for x, c := range result {
		if !isTTY {
			fmt.Fprintf(out, "%s\n", c.Login)
		} else {
			// TODO might regret pre-colorizing this
			logins = append(logins, getColor(x)(c.Login))
		}
	}

	if !isTTY {
		return nil
	}

	lines := []string{}

	thankLines := strings.Split(thankYou, "\n")
	for x, tl := range thankLines {
		lines = append(lines, getColor(x)(tl))
	}
	lines = append(lines, "")
	for _, l := range logins {
		lines = append(lines, fmt.Sprintf("%s", l))
	}

	w, h, err := terminal.GetSize(int(outFile.Fd()))
	if err != nil {
		return err
	}
	fmt.Println(w, h)

	loop := true
	startx := h - 1
	li := 0

	// startx is when to start pulling from the lines array

	for loop {
		clear()
		for x := 0; x < h; x++ {
			if x == startx {
				// pull line from lines; how to tell what index?
				for y := 0; y < li+1; y++ {
					fmt.Fprintf(out, "%s\n", lines[y])
				}
				li += 1
			} else {
				fmt.Fprintf(out, "\n")
			}
		}
		startx -= 1
		if startx == 0 {
			loop = false
		}
		time.Sleep(1000 * time.Millisecond)
	}
	//for i := 0; i < h; i++ {
	//	fmt.Println()
	//	if i == h/2 {
	//		for j, t := range thankLines {
	//			fmt.Fprintf(out, "%s\n", getColor(j)(t))
	//		}
	//		i += len(thankLines)
	//	}
	//}

	// TODO try some animation

	// probably: get term height/width
	// write a character grid
	// flush
	// loop

	//

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

func clear() {
	value, ok := clears[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                           //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

//func main() {
//    fmt.Println("I will clean the screen in 2 seconds!")
//    time.Sleep(2 * time.Second)
//    CallClear()
//    fmt.Println("I'm alone...")
//}
