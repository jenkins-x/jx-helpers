module github.com/jenkins-x/jx-helpers/v3

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.5.2
	github.com/google/uuid v1.1.1
	github.com/jenkins-x/go-scm v1.5.164
	github.com/jenkins-x/jx-api/v3 v3.0.1
	github.com/jenkins-x/jx-kube-client/v3 v3.0.1
	github.com/jenkins-x/jx-logging/v3 v3.0.2
	github.com/jenkins-x/jx-secret v0.0.164 // indirect
	github.com/magiconair/properties v1.8.0
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday v1.5.2
	github.com/satori/go.uuid v1.2.1-0.20180103174451-36e9d2ebbde5
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stoewer/go-strcase v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/xeipuuv/gojsonschema v1.2.0
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/helm v2.16.12+incompatible
	k8s.io/kubernetes v1.14.7
	sigs.k8s.io/kustomize/kyaml v0.6.1
	sigs.k8s.io/yaml v1.2.0
)

go 1.15
