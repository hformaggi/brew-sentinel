package main

import (
	"brew-sentinel/internal/brew"
	"brew-sentinel/internal/i18n"
	"brew-sentinel/internal/tui"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
)

var version = "dev"

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	flags := flag.NewFlagSet("brew-sentinel", flag.ContinueOnError)
	language := flags.String("lang", "", "Interface language: system, en or pt (default: system)")
	jsonOutput := flags.Bool("json", false, "Print installed package inventory as JSON")
	showVersion := flags.Bool("version", false, "Print version and exit")
	if e := flags.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Unexpected positional arguments")
		return 2
	}
	if *showVersion {
		fmt.Fprintln(os.Stdout, "brew-sentinel", version)
		return 0
	}
	if _, e := i18n.Preference(*language); e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 2
	}
	preference := *language
	if preference == "" {
		if path, e := i18n.ConfigPath(); e == nil {
			preference, _ = i18n.Load(path)
		}
	}
	system := "en"
	if preference == "system" || preference == "" {
		system = i18n.System()
	}
	tr := i18n.New(i18n.Resolve(preference, system)).T
	c, e := brew.New()
	if e != nil {
		fmt.Fprintln(os.Stderr, tr("Homebrew not found:"), e)
		return 1
	}
	if *jsonOutput {
		s, e := c.Inventory(context.Background())
		if e != nil {
			fmt.Fprintln(os.Stderr, i18n.ErrorText(e, tr))
			return 1
		}
		if e = json.NewEncoder(os.Stdout).Encode(s); e != nil {
			fmt.Fprintln(os.Stderr, i18n.ErrorText(e, tr))
			return 1
		}
		return 0
	}
	u := tui.NewWithOptions(context.Background(), c, tui.Options{Language: *language})
	if e := u.Run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	return 0
}
