package scmhelpers

import (
	"github.com/jenkins-x/go-scm/scm"
)

// ContainsLabel returns true if the given labels contains the given label name
func ContainsLabel(labels []*scm.Label, label string) bool {
	for _, l := range labels {
		if l != nil && l.Name == label {
			return true
		}
	}
	return false
}
