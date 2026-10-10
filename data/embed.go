// Package data holds every game number as files you can edit.
// The files are built into the program, so the browser never has to fetch them.
package data

import "embed"

// Files is every data file in this folder and its subfolders.
//
//go:embed *.json rules/*.json
var Files embed.FS
