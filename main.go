package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/frou/stdext"
)

var (
	flagBookmarksPath = flag.String("bookmarks",
		platformDefaultBookmarksPath(),
		"path to the SourceTree bookmarks file")

	flagDefaultRebase = flag.Bool("default-rebase",
		true,
		"whether to define the myrepos update command to rebase by default")

	flagConfigRelativePaths = flag.Bool("config-relative-repo-paths",
		true,
		"whether to define the myrepos repo paths relative to the config file")

	flagOutputPath = flag.String("o",
		".mrconfig",
		"path to write myrepos config file to")

	errRepoHasNoOriginRemote = errors.New("git repo has no origin remote")

	supportedRepoTypes = []string{"git"}
)

func main() {
	flag.Parse()
	stdext.Exit(run())
}

func platformDefaultBookmarksPath() string {
	switch runtime.GOOS {
	case "darwin":
		// TODO(DH): Have default Mac path here.
		return "Bookmarks.xml"
	case "windows":
		return os.ExpandEnv(
			`${LOCALAPPDATA}\Atlassian\SourceTree\Bookmarks.xml`)
	default:
		return "path/to/bookmarks"
	}
}

func run() error {
	marks, err := decodeBookmarksFile(*flagBookmarksPath)
	if err != nil {
		return err
	}

	mrFile, err := os.OpenFile(*flagOutputPath,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		stdext.OwnerWritableReg)
	if err != nil {
		return err
	}
	defer mrFile.Close()
	mrFileAbsPath, err := filepath.Abs(*flagOutputPath)
	if err != nil {
		return err
	}

	fmt.Fprintln(mrFile, "# This file was generated from SourceTree bookmarks",
		"by the", stdext.ExecutableBasename(), "command")
	fmt.Fprintln(mrFile)

	if *flagDefaultRebase {
		writeConfigSection(mrFile,
			"DEFAULT",
			"update = git pull --rebase")
	}

	for _, m := range marks {
		repoType := strings.ToLower(m.RepoType)
		repoSupported := false
		for _, suppType := range supportedRepoTypes {
			if repoType == strings.ToLower(suppType) {
				repoSupported = true
				break
			}
		}
		if !repoSupported {
			return fmt.Errorf("Unsupported repo type: %v", repoType)
		}

		repoPathInConfig := m.Path
		if *flagConfigRelativePaths {
			var err error
			repoPathInConfig, err = filepath.Rel(
				filepath.Dir(mrFileAbsPath),
				m.Path)
			if err != nil {
				return err
			}
		}

		originUrl, err := gitOriginFetchURLForRepo(m.Path)
		if err != nil {
			if err != errRepoHasNoOriginRemote {
				return err
			}
			writeConfigSection(mrFile, repoPathInConfig)
		} else {
			writeConfigSection(mrFile,
				repoPathInConfig,
				"checkout = git clone "+originUrl)

		}
	}

	return nil
}

func writeConfigSection(w io.Writer, name string, lines ...string) {
	fmt.Fprintf(w, "[%v]\n", name)
	for _, l := range lines {
		fmt.Fprintln(w, l)
	}
	fmt.Fprintln(w)
}

func gitOriginFetchURLForRepo(repoPath string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath

	url, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			errStr := string(bytes.TrimSpace(ee.Stderr))
			if errStr == "fatal: No such remote 'origin'" {
				return "", errRepoHasNoOriginRemote
			}
		}
		return "", err
	}
	return string(bytes.TrimSpace(url)), nil
}
