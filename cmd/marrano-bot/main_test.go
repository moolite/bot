package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestVersionNotEmpty(t *testing.T) {
	is := is.New(t)
	is.True(version != "")
}

func TestFlagDefaults(t *testing.T) {
	is := is.New(t)
	is.Equal(flagHelp, false)
	is.Equal(flagDebug, false)
	is.Equal(flagInit, false)
	is.Equal(flagDump, false)
	is.Equal(flagExportDB, false)
	is.Equal(flagExportDBPath, "")
	is.Equal(flagSyncMediaFolder, "")
	is.Equal(flagImportHistory, "")
}
