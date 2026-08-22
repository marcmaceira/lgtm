package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"lgtm/internal/config"
	"lgtm/internal/prompt"
)

func runInit(args []string) int {
	dir, err := config.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}

	if _, err := os.Stat(dir + "/config.json"); err == nil {
		fmt.Fprintf(os.Stderr, "lgtm: %s/config.json already exists, leaving it alone\n", dir)
	} else if err := config.Save(config.Default()); err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: writing config:", err)
		return 1
	} else {
		fmt.Printf("lgtm: wrote %s/config.json\n", dir)
	}

	if err := prompt.WriteDefaults(dir); err != nil {
		fmt.Fprintln(os.Stderr, "lgtm: writing templates:", err)
		return 1
	}
	fmt.Printf("lgtm: wrote %s/commit.tmpl and %s/pr.tmpl\n", dir, dir)
	return 0
}

func runConfig(args []string) int {
	fs := flag.NewFlagSet("lgtm config", flag.ContinueOnError)
	commitInstructions := fs.String("commit-instructions", "", "replace commit generation instructions (empty clears them)")
	prInstructions := fs.String("pr-instructions", "", "replace PR generation instructions (empty clears them)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}

	changed := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "commit-instructions":
			cfg.CommitInstructions = *commitInstructions
			changed = true
		case "pr-instructions":
			cfg.PRInstructions = *prInstructions
			changed = true
		}
	})
	if changed {
		if err := config.Save(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "lgtm: writing config:", err)
			return 1
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "lgtm:", err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}
