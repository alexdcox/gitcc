package main

import (
	"os"
	"gopkg.in/urfave/cli.v2"
	"fmt"
	"net/http"
	"github.com/Jeffail/gabs"
	"os/exec"
	"path/filepath"
	"bytes"
	"strings"
)

func main() {
	app := &cli.App{
		Name:  "gitcc",
		Usage: "Clone all github repositories for a given user",
		UsageText: "gitcc [-l LANGUAGE] user",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "language",
				Aliases: []string{"l"},
				Usage:   "Filter retrieved repos by `LANGUAGE`",
			},
		},
		Action: func(c *cli.Context) (err error) {
			switch {
			case os.Getenv("GOPATH") == "",
				!c.Args().Present():
				return cli.ShowAppHelp(c)
			}

			user := c.Args().Get(0)
			user = strings.Replace(user, "github.com/", "", 1)
			fmt.Println("Fetching all github repos for user: " + user)

			rsp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s/repos", user))
			if err != nil {
				return
			}

			jsn, err := gabs.ParseJSONBuffer(rsp.Body)
			if err != nil {
				return
			}

			values, err := jsn.Children()
			if err != nil {
				return
			}

			for _, v := range values {
				repo := strings.Split(v.Path("full_name").Data().(string), "/")[1]
				repoPath := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", user, repo)

				fmt.Printf("• %s ", repo)

				filterLanguage := strings.ToLower(c.String("language"))
				repoLanguage, _ := v.Path("language").Data().(string)
				repoLanguage = strings.ToLower(repoLanguage)
				if filterLanguage != "" && filterLanguage != repoLanguage {
					fmt.Println("SKIPPED (not the specified language)")
					continue
				}

				if _, err := os.Stat(repoPath); os.IsNotExist(err) {
					userRepoDir := filepath.Dir(repoPath)
					if err := os.MkdirAll(userRepoDir, os.ModePerm); err != nil {
						fmt.Println("FAILED (to create user repo directory)")
						continue
					}
					stderr := new(bytes.Buffer)
					cmd := cmd(fmt.Sprintf("git clone https://github.com/%s/%s.git", user, repo), userRepoDir)
					cmd.Stderr = stderr
					if err := cmd.Run(); err != nil {
						fmt.Println("FAILED (to clone new repo)\n", stderr)
					} else {
						fmt.Println("OK!")
					}
				} else {
					switch {
					case !isGitRepoInitialised(repoPath):
						fmt.Println("SKIPPED (dir exists, not a repo)")
						continue
					case doesGitRepoHaveChanges(repoPath):
						fmt.Println("SKIPPED (repo has working changes)")
						continue
					case !isGitRepoOnMasterBranch(repoPath):
						fmt.Println("SKIPPED (repo not on master branch)")
						continue
					}

					stderr := new(bytes.Buffer)
					cmd := exec.Command("git", "pull", "--ff-only", "origin", "master")
					cmd.Dir = repoPath
					cmd.Stderr = stderr
					if err := cmd.Run(); err != nil {
						fmt.Println("FAILED (to pull repo updates)\n", stderr)
					} else {
						fmt.Println("OK!")
					}
				}
			}

			fmt.Println("Done!")

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

func isGitRepoInitialised(dir string) bool {
	return cmd("git status", dir).Run() == nil
}

func doesGitRepoHaveChanges(dir string) bool {
	return cmd("git diff-index --quiet HEAD --", dir).Run() != nil
}

func isGitRepoOnMasterBranch(dir string) bool {
	stdout := new(bytes.Buffer)
	cmd := cmd("git rev-parse --abbrev-ref HEAD", dir)
	cmd.Stdout = stdout
	cmd.Run()
	return strings.TrimSpace(stdout.String()) == "master"
}

func cmd(command, dir string) *exec.Cmd {
	segments := strings.Split(command, " ")
	cmd := exec.Command(segments[0], segments[1:]...)
	cmd.Dir = dir
	return cmd
}
