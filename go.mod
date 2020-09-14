module github.com/jenkins-x/jx-helpers

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.3.0
	github.com/google/uuid v1.1.1
	github.com/jenkins-x/go-scm v1.5.165
	github.com/jenkins-x/jx-api v0.0.17
	github.com/jenkins-x/jx-kube-client v0.0.8
	github.com/jenkins-x/jx-logging v0.0.11
	github.com/magiconair/properties v1.8.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/pborman/uuid v1.2.0
	github.com/petergtz/pegomock v2.7.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday v1.5.2
	github.com/satori/go.uuid v1.2.1-0.20180103174451-36e9d2ebbde5
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stoewer/go-strcase v1.2.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v0.17.6
	k8s.io/helm v2.16.10+incompatible
	k8s.io/kubernetes v1.14.7
	sigs.k8s.io/kustomize/kyaml v0.6.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	//golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // pinned to release-branch.go1.13
	//golang.org/x/tools => golang.org/x/tools v0.0.0-20190821162956-65e3620a7ae7 // pinned to release-branch.go1.13
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.7
)

go 1.13
