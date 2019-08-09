package reportoutput

import (
	"github.com/fusor/cpma/pkg/transform/cluster"
	"github.com/sirupsen/logrus"
)

// ReportOutput holds a collection of reports to be written to file
type ReportOutput struct {
	ClusterReport    cluster.Report    `json:"cluster,omitempty"`
	ComponentReports []ComponentReport `json:"components,omitempty"`
}

// ComponentReport holds a collection of ocp3 config reports
type ComponentReport struct {
	Component string   `json:"component"`
	Reports   []Report `json:"reports"`
}

// Report of OCP 4 component configuration compatibility
type Report struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Supported  bool   `json:"supported"`
	Confidence int    `json:"confidence"`
	Comment    string `json:"comment"`
}

// DumpReports creates OCDs files
func DumpReports(r ReportOutput) {
	// reportOutputFormat := env.Config().GetString("OutputFormat")
	reportOutputFormat := "all"

	switch reportOutputFormat {
	case "json":
		jsonOutput(r)
	case "html":
		htmlOutput(r)
	case "all":
		jsonOutput(r)
		htmlOutput(r)
	default:
		logrus.Error("This format type is not supported")
	}
}