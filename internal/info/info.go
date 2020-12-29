package info

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
)

// App contains information about the running application such as name and version
type App struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Instance contains information about the current instance of the running appplication
type Instance struct {
	Name string `json:"name"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// NewApp transforms application information into a struct
func NewApp(info []byte) (App, error) {
	app := &App{}
	err := json.Unmarshal(info, app)
	return *app, err
}

// InstanceInfo retrieves information about the current application instance
func InstanceInfo() Instance {
	val, err := os.Hostname()
	if err != nil {
		fmt.Println("HOSTNAME not found, generatic")
	}
	return Instance{
		Name: val,
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}
