package info

import (
	"os"
	"runtime"
	"time"
)

var version string
var commitHash string
var buildDate string

// App contains information about the running application such as name and version
type App struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Hash      string `json:"hash"`
	BuildDate string `json:"buildDate"`
}

// Instance contains information about the current instance of the running appplication
type Instance struct {
	Hostname string `json:"Hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
}

// AppInfo transforms application information into a struct
func AppInfo() App {
	app := App{
		Name:      "scratch-post",
		Version:   getVersion(),
		Hash:      getCommitHash(),
		BuildDate: getBuildDate(),
	}
	return app
}

// InstanceInfo retrieves information about the current application instance
func InstanceInfo() Instance {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "scratch-post"
	}
	return Instance{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}
}

func getVersion() string {
	if version == "" {
		return "dev"
	}
	return version
}

func getCommitHash() string {
	if commitHash == "" {
		return "dev"
	}
	return commitHash
}

func getBuildDate() string {
	if buildDate == "" {
		return time.Now().Format(time.RFC3339)
	}
	return buildDate
}
