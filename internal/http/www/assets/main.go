// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package main provides an esbuild pipeline to minify and bundle static assets used by
// the SparkleMuffin Web application.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	cssResult := api.Build(api.BuildOptions{
		EntryPoints: []string{
			"css/www.css",
		},
		Outfile:           "../static/www.min.css",
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Write:             true,
		LogLevel:          api.LogLevelInfo,
		Loader: map[string]api.Loader{
			".css":   api.LoaderCSS,
			".ttf":   api.LoaderFile,
			".woff2": api.LoaderFile,
		},
	})
	if len(cssResult.Errors) > 0 {
		errors := make([]string, len(cssResult.Errors))
		for i, err := range cssResult.Errors {
			errors[i] = err.Text
		}
		log.Fatalf("failed to build CSS: %s\n", strings.Join(errors, ", "))
	}

	if err := copyFile("node_modules/awesomplete/awesomplete.min.js", "../static/awesomplete.min.js"); err != nil {
		log.Fatal(err)
	}
	if err := copyFile("node_modules/easymde/dist/easymde.min.js", "../static/easymde.min.js"); err != nil {
		log.Fatal(err)
	}
	if err := copyFiles("favicons", "../static"); err != nil {
		log.Fatal(err)
	}
}

func copyFile(src, dest string) error {
	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", src, err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", src, dest, err)
	}

	log.Printf("copied %s to %s\n", src, dest)
	return nil
}

func copyFiles(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk %s: %w", srcDir, err)
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}
		dstPath := filepath.Join(dstDir, relPath)

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
		}

		return copyFile(path, dstPath)
	})
}
