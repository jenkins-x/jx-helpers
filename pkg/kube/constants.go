package kube

const (
	// ServiceJenkins is the name of the Jenkins Service
	ServiceJenkins = "jenkins"

	// ServiceChartMuseum the service name of the Helm ChartMuseum service
	ServiceChartMuseum = "jenkins-x-chartmuseum"

	// SecretJenkinsChartMuseum the chart museum secret
	SecretJenkinsChartMuseum = "jenkins-x-chartmuseum"

	// SecretBucketRepo the bucket repo secret if using it as a chart repositoru
	SecretBucketRepo = "jenkins-x-bucketrepo"

	// LocalHelmRepoName is the default name of the local chart repository where CI/CD releases go to
	LocalHelmRepoName = "releases"

	// LabelTeam indicates the team name an environment belongs to
	LabelTeam = "team"

	// LabelEnvironment indicates the name of the environment
	LabelEnvironment = "env"

	// LabelValueDevEnvironment is the value of the LabelTeam label for Development environments (system namespace)
	LabelValueDevEnvironment = "dev"

	// AnnotationHost used to indicate the host if using NodePort Ingress resources on premise without a LoadBalancer
	AnnotationHost = "jenkins.io/host"

	// SecretBasicAuth the name for the Jenkins X basic auth secret
	SecretBasicAuth = "jx-basic-auth" // #nosec
)
