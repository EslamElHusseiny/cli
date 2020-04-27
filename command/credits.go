package command

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
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
	rand.Seed(time.Now().UnixNano())

	clears = make(map[string]func()) //Initialize it
	clears["darwin"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
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
	Args:  cobra.MaximumNArgs(1),
	RunE:  credits,
}

func credits(cmd *cobra.Command, args []string) error {
	ctx := contextForCommand(cmd)
	client, err := apiClientForContext(ctx)
	if err != nil {
		return err
	}

	owner := "cli"
	repo := "cli"
	if len(args) > 0 {
		parts := strings.SplitN(args[0], "/", 2)
		owner = parts[0]
		repo = parts[1]
	}

	type Contributor struct {
		// TODO silly idea: render their avatar. maybe in a cute grid?
		Login string
	}

	type Result []Contributor

	result := Result{}
	body := bytes.NewBufferString("")
	path := fmt.Sprintf("repos/%s/%s/contributors", owner, repo)

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
	lines = append(lines, "( <3 press ctrl-c to quit <3 )")

	termWidth, termHeight, err := terminal.GetSize(int(outFile.Fd()))
	if err != nil {
		return err
	}

	margin := termWidth / 3

	starLinesLeft := []string{}
	for x := 0; x < len(lines); x++ {
		line := ""
		starChance := 0.1
		for y := 0; y < margin; y++ {
			chance := rand.Float64()
			if chance <= starChance {
				charRoll := rand.Float64()
				switch {
				case charRoll < 0.3:
					line += "."
				case charRoll > 0.3 && charRoll < 0.6:
					line += "+"
				default:
					line += "*"
				}
			} else {
				line += " "
			}
		}
		starLinesLeft = append(starLinesLeft, line)
	}

	starLinesRight := []string{}
	for x := 0; x < len(lines); x++ {
		line := ""
		lineWidth := termWidth - (margin + len(lines[x]))
		starChance := 0.1
		for y := 0; y < lineWidth; y++ {
			chance := rand.Float64()
			if chance <= starChance {
				if rand.Float64() < 0.5 {
					line += "*"
				} else {
					line += "."
				}
			} else {
				line += " "
			}
		}
		starLinesRight = append(starLinesRight, line)
	}

	loop := true
	startx := termHeight - 1
	li := 0

	for loop {
		clear()
		for x := 0; x < termHeight; x++ {
			if x == startx || startx < 0 {
				starty := 0
				if startx < 0 {
					starty = int(math.Abs(float64(startx)))
				}
				for y := starty; y < li+1; y++ {
					if y >= len(lines) {
						continue
					}
					starLineLeft := starLinesLeft[y]
					starLinesLeft[y] = twinkle(starLineLeft)
					starLineRight := starLinesRight[y]
					starLinesRight[y] = twinkle(starLineRight)
					fmt.Fprintf(out, "%s %s %s\n", starLineLeft, lines[y], starLineRight)
				}
				li += 1
				x += li
			} else {
				fmt.Fprintf(out, "\n")
			}
		}
		if li < len(lines) {
			startx -= 1
		}
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

func twinkle(starLine string) string {
	starLine = strings.ReplaceAll(starLine, ".", "P")
	starLine = strings.ReplaceAll(starLine, "+", "A")
	starLine = strings.ReplaceAll(starLine, "*", ".")
	starLine = strings.ReplaceAll(starLine, "P", "+")
	starLine = strings.ReplaceAll(starLine, "A", "*")
	return starLine
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
