package scmhelpers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"

	"github.com/spf13/cobra"
)

type PullRequestOptions struct {
	Options

	// Number the Pull Request number
	Number int

	// IgnoreMissingPullRequest does not return an error if no pull request could be found
	IgnoreMissingPullRequest bool
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
		return fmt.Errorf("failed to validate repository options: %w", err)
	}

	if o.Number <= 0 {
		o.Number, err = FindPullRequestFromEnvironment()
		if err != nil {
			if o.IgnoreMissingPullRequest {
				log.Logger().Warnf("could not find Pull Request number from environment. Assuming main branch instead")
				return nil
			}
			return fmt.Errorf("failed to get PullRequest from environment. Try supplying option: --pr: %w", err)
		}

		if o.Number <= 0 && !o.IgnoreMissingPullRequest {
			return options.MissingOption("pr")
		}
	}
	return nil
}

// FindPullRequestFromEnvironment returns the PullRequest number by looking for common Jenkins X environment variables
func FindPullRequestFromEnvironment() (int, error) {
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
	if prName == "" {
		return 0, nil
	}
	answer, err := strconv.Atoi(prName)
	if err != nil {
		return 0, fmt.Errorf("unable to convert PR "+prName+" to a number from env var %s"+": %w", envVar, err)
	}
	return answer, nil
}

// DiscoverPullRequest discovers the pull request for the current number
func (o *PullRequestOptions) DiscoverPullRequest() (*scm.PullRequest, error) {
	ctx := context.Background()
	pr, _, err := o.ScmClient.PullRequests.Find(ctx, o.FullRepositoryName, o.Number)
	if err != nil {
		return nil, fmt.Errorf("failed to find PR %d in repo %s: %w", o.Number, o.Repository, err)
	}
	return pr, nil
}
