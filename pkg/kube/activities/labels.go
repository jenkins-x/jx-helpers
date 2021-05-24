package activities

import (
	v1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
)

var (
	OwnerLabels   = []string{"owner", "lighthouse.jenkins-x.io/refs.org"}
	RepoLabels    = []string{"repository", "lighthouse.jenkins-x.io/refs.repo"}
	BranchLabels  = []string{"branch", "lighthouse.jenkins-x.io/branch"}
	BuildLabels   = []string{"build", "lighthouse.jenkins-x.io/buildNum"}
	ContextLabels = []string{"context", "lighthouse.jenkins-x.io/context"}
)

// GetLabel returns the first label value for the given strings
func GetLabel(m map[string]string, labels []string) string {
	if m == nil {
		return ""
	}
	for _, l := range labels {
		value := m[l]
		if value != "" {
			return value
		}
	}
	return ""
}

// DefaultValues default missing values from the lighthouse labels
func DefaultValues(a *v1.PipelineActivity) {
	labels := a.Labels
	if labels != nil {
		if a.Spec.GitOwner == "" {
			a.Spec.GitOwner = GetLabel(labels, OwnerLabels)
		}
		if a.Spec.GitRepository == "" {
			a.Spec.GitRepository = GetLabel(labels, RepoLabels)
		}
		if a.Spec.GitBranch == "" {
			a.Spec.GitBranch = GetLabel(labels, BranchLabels)
		}
		if a.Spec.Context == "" {
			a.Spec.Context = GetLabel(labels, ContextLabels)
		}
		if a.Spec.Build == "" {
			a.Spec.Build = GetLabel(labels, BuildLabels)
		}
	}
	if a.Spec.StartedTimestamp == nil {
		for _, s := range a.Spec.Steps {
			if s.Stage != nil {
				a.Spec.StartedTimestamp = s.Stage.StartedTimestamp
			} else if s.Promote != nil {
				a.Spec.StartedTimestamp = s.Promote.StartedTimestamp
			} else if s.Preview != nil {
				a.Spec.StartedTimestamp = s.Preview.StartedTimestamp
			}
			if a.Spec.StartedTimestamp != nil {
				break
			}
		}
	}
	if string(a.Spec.Status) == "" {
		// lets default the status to the last step if its missing
		for i := len(a.Spec.Steps) - 1; i > -0; i-- {
			s := a.Spec.Steps[i]
			status := v1.ActivityStatusTypeNone
			if s.Stage != nil {
				status = s.Stage.Status
			} else if s.Promote != nil {
				status = s.Promote.Status
			} else if s.Preview != nil {
				status = s.Preview.Status
			}

			if string(status) != "" {
				a.Spec.Status = status
				break
			}
		}
	}
}
