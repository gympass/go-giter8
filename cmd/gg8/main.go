package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/gympass/go-giter8/lexer"
	"github.com/gympass/go-giter8/props"
	"github.com/gympass/go-giter8/render"
)

func findGit() (bool, string) {
	p, err := exec.LookPath("git")
	if err != nil {
		return false, ""
	}
	return true, p
}

func errorf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func fatalf(format string, a ...interface{}) {
	errorf(format, a...)
	os.Exit(1)
}

func printf(format string, a ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(format, a...))
}

var githubRepositoryRegexp = regexp.MustCompile(`(?i)^[a-z\d](?:[a-z\d]|-([a-z\d])){0,38}/[a-z0-9\-._]+$`)
var helpRegexp = regexp.MustCompile(`(?i)^((-(-)?)?/?(help|usage))`)

func usage() {
	help := []string{
		"gg8 (go-giter8) - giter8 alternative in Go",
		"",
		"Usage",
		"gg8 REPOSITORY TARGET [-- [option=value]]",
		"",
		"REPOSITORY - Either username/repo for GitHub repositories, or the",
		"             full repository HTTPS/SSH path to clone",
		"TARGET     - Directory to apply template to",
		"",
		"Using option=value",
		"When using option=value, gg8 will not ask for options, and will merge",
		"all provided options into options provided by the repository, ",
		"overwriting existing options.",
	}

	for _, s := range help {
		fmt.Println(s)
	}
}

type TemplateMeta struct {
	HasProperties bool
	Root          string
}

const propsFile = "default.properties"

func detectTemplateMeta(root string) (result TemplateMeta) {
	s, err := os.Stat(path.Join(root, propsFile))
	if err == nil && !s.IsDir() {
		result.HasProperties = true
		result.Root = root
		return
	}

	// Do we have a standard g8 structure?
	result.Root = path.Join(root, "src", "main", "g8")
	s, err = os.Stat(result.Root)
	if err == nil && s.IsDir() {
		s, err = os.Stat(path.Join(result.Root, propsFile))
		if err == nil && !s.IsDir() {
			result.HasProperties = true
		}
	}
	return
}

func clone(gitPath, repo, target string) (bool, string) {
	cmd := exec.Command(gitPath, "clone", repo, target)
	if err := cmd.Start(); err != nil {
		return false, fmt.Sprintf("Error starting process: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return false, fmt.Sprintf("Process exited with non-zero status code %d:\n\n%s", err.ExitCode(), string(err.Stderr))
		}
	}

	return true, ""
}

func main() {
	hasGit, gitPath := findGit()
	if !hasGit {
		fatalf("Could not find `git' in your system. Please ensure it is installed and available through the PATH variable.")
		os.Exit(1)
	}

	if len(os.Args) == 0 {
		usage()
		os.Exit(1)
	}

	if len(os.Args) == 1 && helpRegexp.MatchString(os.Args[0]) {
		usage()
		os.Exit(0)
	}

	repo := ""
	target := ""
	takingOpts := false
	var options props.Pairs

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if arg == "--" {
			if repo == "" {
				fatalf("Found `--' before repository argument. Run gg8 with --help for further information")
			}
			if target == "" {
				fatalf("Found `--' before destination argument. Run gg8 with --help for further information")
			}
			takingOpts = true
			continue
		}
		if repo == "" {
			if githubRepositoryRegexp.MatchString(arg) {
				suffix := ""
				if !strings.HasSuffix(arg, ".git") {
					suffix = ".git"
				}
				repo = fmt.Sprintf("https://github.com/%s%s", arg, suffix)
			} else {
				repo = arg
			}
			continue
		}

		if !takingOpts && repo != "" {
			target = arg
			continue
		}

		if !takingOpts && repo != "" && target != "" {
			fatalf("Unexpected param `%s`. Run gg8 with --help for further information", arg)
		}

		indexOf := strings.Index(arg, "=")
		if indexOf == -1 {
			fatalf("Invalid argument `%s': Arguments must be declared as <key> = <value>", arg)
		}
		options = append(options, props.Pair{
			K: arg[0:indexOf],
			V: arg[indexOf+1:],
		})
	}

	if repo == "" {
		usage()
		os.Exit(1)
	}

	if target == "" {
		usage()
		os.Exit(1)
	}

	if newTarget, err := filepath.Abs(target); err != nil {
		fatalf("Error calculating absolute path for `%s': %s", target, err)
	} else {
		target = newTarget
	}

	// create clone destination
	cloneDir, err := os.MkdirTemp("", "gg8")
	if err != nil {
		fatalf("Error creating temporary directory: %s", err)
	}

	printf("Cloning %s...", repo)
	ok, errStr := clone(gitPath, repo, cloneDir)
	if !ok {
		fatalf("%s", errStr)
	}

	if err := os.RemoveAll(path.Join(cloneDir, ".git")); err != nil {
		fatalf("Error cleaning up cloned template: %s", err)
	}

	templateMeta := detectTemplateMeta(cloneDir)
	var currentProps = props.Pairs{{K: "name", V: filepath.Base(target)}}

	if templateMeta.HasProperties && len(options) == 0 {
		printf("Preparing template:")

		rawProps, err := os.ReadFile(path.Join(templateMeta.Root, propsFile))
		if err != nil {
			fatalf("Error reading %s: %s", propsFile, err)
		}

		allProps, err := props.ParseProperties(string(rawProps))
		if err != nil {
			fatalf("Error parsing %s: %s", propsFile, err)
		}
		allProps.Merge(currentProps)

		for _, p := range allProps {
			r := render.NewExecutor(currentProps)
			propAST, err := lexer.Tokenize(p.V)
			if err != nil {
				fatalf("Error parsing property %s: %s", p.K, err)
			}
			computedValue, err := r.Exec(propAST)
			if err != nil {
				fatalf("Error populating property %s: %s", p.K, err)
			}
			prompt := promptui.Prompt{
				Default: computedValue,
				Label:   p.K,
			}
			promptResult, err := prompt.Run()
			if err != nil {
				fatalf("Error executing prompt: %s", err)
			}
			currentProps = append(currentProps, props.Pair{K: p.K, V: promptResult})
		}
	} else if templateMeta.HasProperties && len(options) != 1 {
		currentProps = options
	}
	printf("\nRendering template to %s", target)
	err = render.TemplateDirectory(currentProps, templateMeta.Root, target)
	if err != nil {
		fatalf("Error rendering directory template: %s", err)
	}
}
