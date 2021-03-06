package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/deis/deisrel/actions"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	ghTkn := os.Getenv("GITHUB_TOKEN")
	if ghTkn == "" {
		log.Fatalf("'GITHUB_TOKEN' env var required")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghTkn})
	cl := oauth2.NewClient(oauth2.NoContext, ts)
	ghClient := github.NewClient(cl)

	app := cli.NewApp()
	app.Name = "deisrel"
	app.Usage = "Utilities for releasing a new Deis version"
	app.Commands = []cli.Command{
		cli.Command{
			Name: "git",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "shas",
					Action: actions.GetShas(ghClient),
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  actions.ShortFlag,
							Usage: "Whether to show short 7 character shas",
						},
					},
				},
			},
		},
		cli.Command{
			Name: "generate-changelog",
			Action: actions.GenerateChangelog(ghClient, os.Stdout),
		},
		cli.Command{
			Name: "helm-params",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  actions.TagFlag,
					Value: "",
					Usage: "The Docker tag to apply for all images. If empty, defaults to 'git-$SHORT_SHA' for each respective component",
				},
				cli.StringFlag{
					Name:  actions.PullPolicyFlag,
					Value: "IfNotPresent",
					Usage: "The 'imagePullPolicy' value to set on each image",
				},
				cli.StringFlag{
					Name:  actions.OrgFlag,
					Value: "deis",
					Usage: "The docker repository organization to set on each image",
				},
			},
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "e2e",
					Action: actions.HelmGenerateE2E(ghClient),
				},
				cli.Command{
					Name:   "workflow",
					Action: actions.HelmGenerateWorkflow(ghClient),
				},
			},
		},
		cli.Command{
			Name: "helm-stage",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  actions.RepoFlag,
					Value: "charts",
					Usage: "The GitHub repo name to grab chart from",
				},
				cli.StringFlag{
					Name:  actions.RefFlag,
					Value: "",
					Usage: "Optional ref to add to GET request (can be SHA, branch or tag); will be omitted if empty",
				},
				cli.StringFlag{
					Name:  actions.GHOrgFlag,
					Value: "deis",
					Usage: "The GitHub org to find repo in",
				},
			},
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "e2e",
					Action:      actions.HelmStageE2E(ghClient),
					Description: "Stages workflow-dev-e2e into staging, amending with $DEIS_RELEASE_SHORT if defined",
				},
				cli.Command{
					Name:        "workflow",
					Action:      actions.HelmStageWorkflow(ghClient),
					Description: "Stages workflow-dev into staging, amending with $DEIS_RELEASE_SHORT if defined",
				},
			},
			Description: `Stages chart files into staging.
			To amend files with values pertinent to a release, user must export the following env variables:
			$DEIS_RELEASE=<full_semver_release_string>, i.e. 'v1.0.0-alpha1'
			$DEIS_RELEASE_SHORT=<short_form_release_string>, i.e. 'alpha1'`,
		},
	}

	app.Run(os.Args)
}
