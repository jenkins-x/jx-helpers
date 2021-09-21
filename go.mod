module github.com/jenkins-x/jx-helpers/v3

require (
	github.com/Azure/go-autorest/autorest v0.11.19 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.14 // indirect
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fatih/color v1.9.0
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.5.5
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.1.4
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jenkins-x/go-scm v1.10.10
	github.com/jenkins-x/jx-api/v4 v4.1.5
	github.com/jenkins-x/jx-kube-client/v3 v3.0.2
	github.com/jenkins-x/jx-logging/v3 v3.0.6
	github.com/kr/pretty v0.2.1 // indirect
	github.com/magiconair/properties v1.8.5
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday v1.6.0
	github.com/sethvargo/go-envconfig v0.3.5
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.6.2
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubernetes v1.14.7
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
	sigs.k8s.io/kustomize/kyaml v0.10.15
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// helm dependencies
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

	k8s.io/api => k8s.io/api v0.20.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.6
	k8s.io/client-go => k8s.io/client-go v0.20.6
)

go 1.15
