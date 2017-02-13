package version

var (
	GitCommit string = "library-import"
	Version   string = "library-import"
	BuildTime string = "library-import"
)

type VersionOptions struct {
	GitCommit string
	Version   string
	BuildTime string
	GoVersion string
	Os        string
	Arch      string
}
