package helmer

// ChartSummary contains a chart summary
type ChartSummary struct {
	Name         string
	ChartVersion string
	AppVersion   string
	Description  string
}

// ReleaseSummary is the information about a release in Helm
type ReleaseSummary struct {
	ReleaseName   string
	Revision      string
	Updated       string
	Status        string
	ChartFullName string
	Chart         string
	ChartVersion  string
	AppVersion    string
	Namespace     string
}
