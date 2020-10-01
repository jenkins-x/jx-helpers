package testjx

import (
	jenkinsio "github.com/jenkins-x/jx-api/v3/pkg/apis/jenkins.io"
	jenkinsv1 "github.com/jenkins-x/jx-api/v3/pkg/apis/jenkins.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateSourceRepository tests a SourceRepository instance for a test case
func CreateSourceRepository(ns, org, repo, kind, provider string) *jenkinsv1.SourceRepository {
	return &jenkinsv1.SourceRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SourceRepository",
			APIVersion: jenkinsio.GroupName + "/" + jenkinsio.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      org + "-" + repo,
			Namespace: ns,
		},
		Spec: jenkinsv1.SourceRepositorySpec{
			Provider:     provider,
			Org:          org,
			Repo:         repo,
			ProviderKind: kind,
			ProviderName: kind,
		},
	}
}
