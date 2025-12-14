package main

import (
	"floapp/nem12"
	initilize "floapp/standard-library/initialize"
	"os"
	"path/filepath"

	"github.com/beego/beego/v2/core/logs"
)

func main() {
	// Start DB
	if err := initilize.InitDB(); err != nil {
		logs.Error("db init failed: %v", err)
		os.Exit(1)
	}
	// NEM12 File
	inPath := "NEM.txt"
	if len(os.Args) > 1 && os.Args[1] != "" {
		inPath = os.Args[1]
	}

	//Get Env is linux or windows
	cwd, errWd := os.Getwd()
	if errWd == nil {
		if !filepath.IsAbs(inPath) {
			inPath = filepath.Join(cwd, inPath)
		}
	}

	stats, err := nem12.UpsertFromPath(inPath, nem12.GeneratorOptions{BatchSize: 100})
	if err != nil {
		logs.Error("upsert failed: %v", err)
		os.Exit(1)
	}
	logs.Info("upsert ok input=%s records=%d inserts/updates=%d", inPath, stats.Records, stats.Inserts)
}
