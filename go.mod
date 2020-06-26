module github.com/jenkins-x/jx-helpers

require (
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/jenkins-x/jx-logging v0.0.8
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20180103174451-36e9d2ebbde5
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/testify v1.6.0
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092 // indirect
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	google.golang.org/appengine v1.1.0 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.3
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	sigs.k8s.io/yaml v1.2.0

)

replace github.com/heptio/sonobuoy => github.com/jenkins-x/sonobuoy v0.11.7-0.20190318120422-253758214767

replace k8s.io/api => k8s.io/api v0.0.0-20190528110122-9ad12a4af326

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20181128195641-3954d62a524d

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221084156-01f179d85dbc

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190528110200-4f3abb12cae2

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190528110544-fa58353d80f3

replace github.com/sirupsen/logrus => github.com/jtnord/logrus v1.4.2-0.20190423161236-606ffcaf8f5d

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v21.1.0+incompatible

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v10.15.5+incompatible

replace github.com/banzaicloud/bank-vaults => github.com/banzaicloud/bank-vaults v0.0.0-20190508130850-5673d28c46bd

replace github.com/TV4/logrus-stackdriver-formatter => github.com/jenkins-x/logrus-stackdriver-formatter v0.1.1-0.20200408213659-1dcf20c371bb

replace code.gitea.io/sdk/gitea => code.gitea.io/sdk/gitea v0.12.0

go 1.13
