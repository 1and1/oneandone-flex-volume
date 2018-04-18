package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/1and1/oneandone-flex-volume/cmd/oneandone-flex-volume/config"
	"github.com/1and1/oneandone-flex-volume/helper"
	"github.com/1and1/oneandone-flex-volume/pkg/flex"
	"github.com/1and1/oneandone-flex-volume/pkg/oneandone/cloud"
	"github.com/1and1/oneandone-flex-volume/pkg/oneandone/plugin"
	"github.com/golang/glog"
)

// func init() {
// 	flag.Set("logtostderr", "true")
// }

func main() {

	flag.Set("logtostderr", "true")
	flag.Parse()

	// Create the 1&1 manager
	token, err := config.GetOneandoneToken()
	if err != nil {
		glog.Errorf("Error retrieving 1&1 token: %v", err.Error())
		os.Exit(1)
	}
	oneandone, err := cloud.NewOneandoneManager(token)
	if err != nil {
		glog.Errorf("Error creating 1and1 client: %v", err.Error())
		os.Exit(1)
	}

	// create 1&1 flex volume instance
	p := plugin.NewOneandoneVolumePlugin(oneandone)
	// create flex Executor
	manager := flex.NewManager(p, os.Stdout)

	// read arguments
	args := os.Args
	if len(args) < 2 {
		manager.WriteError(fmt.Errorf("flex command argument was not found"))
		os.Exit(1)
	}

	// create flex command based on flags
	fc, err := flex.NewFlexCommand(args)
	helper.DebugFile(fmt.Sprintf("Arguments recieved %s", args))
	if err != nil {
		helper.DebugFile(fmt.Sprintf("COMMAND CREATE error %s", err.Error()))
		manager.WriteError(err)
		os.Exit(1)
	}
	helper.DebugFile(fmt.Sprintf("command recieved %s", fc))

	// execute flex command
	ds, err := manager.ExecuteCommand(fc)

	if err != nil {
		helper.DebugFile(fmt.Sprintf("EXECUTE error %s", err.Error()))
		manager.WriteError(err)
		os.Exit(1)
	}

	// write result to output
	err = manager.WriteDriverStatus(ds)
	if err != nil {
		helper.DebugFile(fmt.Sprintf("DRIVER STATUS error: %s", err.Error()))
		manager.WriteError(err)
		os.Exit(1)
	}
}
