package testjx

import (
	jenkinsio "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io"
	jxCore "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateSourceRepository tests a SourceRepository instance for a test case
func CreateSourceRepository(ns, org, repo, kind, provider string) *jxCore.SourceRepository {
	return &jxCore.SourceRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SourceRepository",
			APIVersion: jenkinsio.GroupName + "/" + jenkinsio.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      org + "-" + repo,
			Namespace: ns,
		},
		Spec: jxCore.SourceRepositorySpec{
			Provider:     provider,
			Org:          org,
			Repo:         repo,
			ProviderKind: kind,
			ProviderName: kind,
		},
	}
}
