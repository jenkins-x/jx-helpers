package pipelinectx

import (
	"context"
	"strconv"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/naming"
	"github.com/jenkins-x/jx-helpers/v3/pkg/options"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/cobra"
)

// Options default options for Jenkins X Pipelines
//
// see: https://jenkins-x.io/v3/develop/pipelines/reference/#parameters-and-environment-variables
type Options struct {
	// AppName the name of the app
	AppName string `env:"APP_NAME"`

	// RepoName is the name of the repository
	RepoName string `env:"REPO_NAME"`

	// RepoOwner is the user/organisation which owns the repository
	RepoOwner string `env:"REPO_OWNER"`

	// BuildNumber is build number
	BuildNumber string `env:"BUILD_NUMBER"`

	// Context is the pipeline context if there are multiple contexts.
	Context string `env:"JOB_NAME"`

	// JobType is the type of job (e.g. presubmit or postsubmit)
	JobType string `env:"JOB_TYPE"`

	// BranchName is the work directory. If not specified a temporary directory is created on startup.
	BranchName string `env:"BRANCH_NAME"`

	// Version the version being build
	Version string `env:"VERSION"`

	// PullRequestNumber the PR number if inside a Pull Request pipeline
	PullRequestNumber int `env:"PULL_NUMBER"`

	// PullSHA the git sha of the Pull Request
	PullSHA string `env:"PULL_PULL_SHA"`

	// ResourceName the unique k8s resource name we can use to create, say, Terraform instances
	ResourceName string

	// ResourceNamePrefix allows a prefix to be prepended to the resource name
	ResourceNamePrefix string

	// Labels labels for the resource
	Labels map[string]string
}

// EnvironmentDefaults processes environment defaults before we parse CLI arguments
func (o *Options) EnvironmentDefaults(ctx context.Context) error {
	err := envconfig.Process(ctx, o)
	if err != nil {
		return errors.Wrapf(err, "failed to process environment options")
	}
	return nil
}

// Validate validates values
func (o *Options) Validate() error {
	if o.AppName == "" {
		o.AppName = o.RepoName
	}
	if o.BuildNumber == "" {
		return options.MissingOption("build-number")
	}

	context := o.Context
	if context != "" {
		context = "-" + context
	}
	prNumber := ""
	pr := ""
	if o.PullRequestNumber > 0 {
		prNumber = strconv.Itoa(o.PullRequestNumber)
		pr = "-pr" + prNumber
	}

	o.ResourceName = o.ResourceNamePrefix + o.RepoName + pr + context + "-" + o.BuildNumber

	if o.Labels == nil {
		o.Labels = map[string]string{}
	}

	if o.Context != "" && o.Labels["context"] == "" {
		o.Labels["context"] = naming.ToValidName(o.Context)
	}
	if o.RepoName != "" && o.Labels["repo"] == "" {
		o.Labels["repo"] = naming.ToValidName(o.RepoName)
	}
	if o.RepoOwner != "" && o.Labels["owner"] == "" {
		o.Labels["owner"] = naming.ToValidName(o.RepoOwner)
	}
	if prNumber != "" && o.Labels["pr"] == "" {
		o.Labels["pr"] = naming.ToValidName("PR-" + prNumber)
	}
	return nil
}

// AddFlags adds CLI flags
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.AppName, "app", "", o.AppName, "the name of the app. Defaults to $APP_NAME")
	cmd.Flags().StringVarP(&o.RepoName, "repo", "", o.RepoName, "the name of the repository. Defaults to $REPO_NAME")
	cmd.Flags().StringVarP(&o.RepoOwner, "owner", "", o.RepoOwner, "the owner of the repository. Defaults to $REPO_OWNER")
	cmd.Flags().StringVarP(&o.Context, "context", "", o.Context, "the pipeline context. Defaults to $JOB_NAME")
	cmd.Flags().StringVarP(&o.JobType, "type", "", o.JobType, "the pipeline type. e.g. presubmit or postsubmit. Defaults to $JOB_TYPE")
	cmd.Flags().StringVarP(&o.BranchName, "branch", "", o.BranchName, "the branch used in the pipeline. Defaults to $BRANCH_NAME")
	cmd.Flags().StringVarP(&o.BuildNumber, "build", "", o.BuildNumber, "the build number. Defaults to $BUILD_NUMBER")
	cmd.Flags().StringVarP(&o.Version, "version", "", o.Version, "the version number. Defaults to $VERSION")
	cmd.Flags().StringVarP(&o.PullSHA, "pull-sha", "", o.PullSHA, "the Pull Request git SHA. Defaults to $PULL_PULL_SHA")
	cmd.Flags().StringVarP(&o.ResourceNamePrefix, "name-prefix", "", o.ResourceNamePrefix, "the resource name prefix")
	cmd.Flags().IntVarP(&o.PullRequestNumber, "pr", "", o.PullRequestNumber, "the Pull Request number. Defaults to $PULL_NUMBER")
}
