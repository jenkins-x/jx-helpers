package scmhelpers

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/pkg/options"
	"github.com/jenkins-x/jx-logging/pkg/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type PullRequestOptions struct {
	Options

	// Number the Pull Request number
	Number int
}

// AddFlags adds CLI flags
func (o *PullRequestOptions) AddFlags(cmd *cobra.Command) {
	o.Options.AddFlags(cmd)

	cmd.Flags().IntVarP(&o.Number, "pr", "", 0, "the Pull Request number. If not specified we detect it via $PULL_NUMBER or $BRANCH_NAME environment variables")

}

// Validate validates the inputs are valid
func (o *PullRequestOptions) Validate() error {
	err := o.Options.Validate()
	if err != nil {
		return errors.Wrapf(err, "failed to validate repository options")
	}

	if o.Number <= 0 {
		envVar := "PULL_NUMBER"
		prName := os.Getenv(envVar)
		if prName == "" {
			envVar = "BRANCH_NAME"
			branchName := strings.ToUpper(os.Getenv(envVar))
			prPrefix := "PR-"
			if strings.HasPrefix(branchName, prPrefix) {
				prName = strings.TrimPrefix(branchName, prPrefix)
			}
		}
		if prName != "" {
			o.Number, err = strconv.Atoi(prName)
			if err != nil {
				log.Logger().Warnf(
					"Unable to convert PR "+prName+" to a number from env var %s", envVar)
			}
		}
		if o.Number <= 0 {
			return options.MissingOption("pr")
		}
	}
	return nil
}

// DiscoverPullRequest discovers the pull request for the current number
func (o *PullRequestOptions) DiscoverPullRequest() (*scm.PullRequest, error) {
	ctx := context.Background()
	pr, _, err := o.ScmClient.PullRequests.Find(ctx, o.FullRepositoryName, o.Number)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find PR %d in repo %s", o.Number, o.Repository)
	}
	return pr, nil
}
