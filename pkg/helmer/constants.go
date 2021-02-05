package helmer

const (
	// ChartFileName file name for a chart
	ChartFileName = "Chart.yaml"
	// RequirementsFileName the file name for helm requirements
	RequirementsFileName = "requirements.yaml"
	// SecretsFileName the file name for secrets
	SecretsFileName = "secrets.yaml"
	// ValuesFileName the file name for values
	ValuesFileName = "values.yaml"
	// ValuesTemplateFileName a templated values.yaml file which can refer to parameter expressions
	ValuesTemplateFileName = "values.tmpl.yaml"
	// TemplatesDirName is the default name for the templates directory
	TemplatesDirName = "templates"

	// ParametersYAMLFile contains logical parameters (values or secrets) which can be fetched from a Secret URL or
	// inlined if not a secret which can be referenced from a 'values.yaml` file via a `{{ .Parameters.foo.bar }}` expression
	ParametersYAMLFile = "parameters.yaml"

	// FakeChartmusuem is the url for the fake chart museum used in tests
	FakeChartmusuem = "http://fake.chartmuseum"

	// DefaultEnvironmentChartDir is the default environment path where charts are stored
	DefaultEnvironmentChartDir = "env"

	//RepoVaultPath is the path to the repo credentials in Vault
	RepoVaultPath = "helm/repos"

	// JX3ChartRepository the default charts repo for the jx3 charts
	JX3ChartRepository = "https://storage.googleapis.com/jenkinsxio/charts"

	// AnnotationChartName stores the chart name
	AnnotationChartName = "jenkins.io/chart"
	// AnnotationAppVersion stores the chart's app version
	AnnotationAppVersion = "jenkins.io/chart-app-version"
	// AnnotationAppDescription stores the chart's app version
	AnnotationAppDescription = "jenkins.io/chart-description"
	// AnnotationAppRepository stores the chart's app repository
	AnnotationAppRepository = "jenkins.io/chart-repository"

	// LabelReleaseName stores the chart release name
	LabelReleaseName = "jenkins.io/chart-release"

	// LabelNamespace stores the chart namespace for cluster wide resources
	LabelNamespace = "jenkins.io/namespace"

	// LabelReleaseChartVersion stores the version of a chart installation in a label
	LabelReleaseChartVersion = "jenkins.io/version"
	// LabelAppName stores the chart's app name
	LabelAppName = "jenkins.io/app-name"
	// LabelAppVersion stores the chart's app version
	LabelAppVersion = "jenkins.io/app-version"

	hookFailed    = "hook-failed"
	hookSucceeded = "hook-succeeded"

	// resourcesSeparator is used to separate multiple objects stored in the same YAML file
	resourcesSeparator = "---"
)

//DefaultValuesTreeIgnores is the default set of ignored files for collapsing the values tree which are used if
// ignores is nil
var DefaultValuesTreeIgnores = []string{
	"templates/*",
}
