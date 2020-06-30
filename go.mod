module github.com/jenkins-x/jx-helpers

require (
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/alecthomas/jsonschema v0.0.0-20200530073317-71f438968921
	github.com/blang/semver v3.5.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.0
	github.com/go-openapi/spec v0.19.7
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/go-cmp v0.3.0
	github.com/imdario/mergo v0.3.9
	github.com/jenkins-x/jx-api v0.0.11
	github.com/jenkins-x/jx-logging v0.0.8
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday v1.5.2
	github.com/satori/go.uuid v1.2.1-0.20180103174451-36e9d2ebbde5
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stoewer/go-strcase v1.2.0
	github.com/stretchr/testify v1.6.1
	github.com/vrischmann/envconfig v1.2.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/tools v0.0.0-20200415034506-5d8e1897c761
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/src-d/go-git.v4 v4.13.1
	k8s.io/api v0.16.5
	k8s.io/apimachinery v0.16.5
	k8s.io/client-go v0.16.5
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	sigs.k8s.io/yaml v1.1.0
)

replace (
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // pinned to release-branch.go1.13
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190821162956-65e3620a7ae7 // pinned to release-branch.go1.13
	k8s.io/api => k8s.io/api v0.16.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190819143637-0dbe462fe92d
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.5
	k8s.io/client-go => k8s.io/client-go v0.16.5
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190819143841-305e1cef1ab1
)

go 1.13
