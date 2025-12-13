// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

// Package main provides an esbuild pipeline to minify and bundle static assets used by
// the SparkleMuffin Web application.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/evanw/esbuild/pkg/api"
)

var cssBuildOptions = api.BuildOptions{
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
}

var jsBuildOptions = api.BuildOptions{
	EntryPoints: []string{
		"js/easymde-init.js",
	},
	Outdir:            "../static",
	Bundle:            true,
	MinifyWhitespace:  true,
	MinifyIdentifiers: true,
	MinifySyntax:      true,
	Write:             true,
	LogLevel:          api.LogLevelInfo,
	OutExtension: map[string]string{
		".js": ".min.js",
	},
}

func main() {
	watchMode := flag.Bool("watch", false, "Watch for changes and rebuild automatically")
	flag.Parse()

	// Copy static assets (JS libraries, favicons)
	copyStaticAssets()

	if *watchMode {
		watchAssets()
	} else {
		buildAssets()
	}
}

func buildAssets() {
	cssResult := api.Build(cssBuildOptions)
	if len(cssResult.Errors) > 0 {
		errors := make([]string, len(cssResult.Errors))
		for i, err := range cssResult.Errors {
			errors[i] = err.Text
		}
		log.Fatalf("esbuild: failed to build CSS assets: %s\n", strings.Join(errors, ", "))
	}

	jsResult := api.Build(jsBuildOptions)
	if len(jsResult.Errors) > 0 {
		errors := make([]string, len(jsResult.Errors))
		for i, err := range jsResult.Errors {
			errors[i] = err.Text
		}
		log.Fatalf("esbuild: failed to build JS assets: %s\n", strings.Join(errors, ", "))
	}
}

func watchAssets() {
	cssCtx, err := api.Context(cssBuildOptions)
	if err != nil {
		log.Fatalf("esbuild: failed to create CSS esbuild context: %s\n", err)
	}

	jsCtx, err := api.Context(jsBuildOptions)
	if err != nil {
		log.Fatalf("esbuild: failed to create JS esbuild context: %s\n", err)
	}

	// Start watching
	if err := cssCtx.Watch(api.WatchOptions{}); err != nil {
		log.Fatalf("esbuild: failed to start CSS watch mode: %s\n", err)
	}
	if err := jsCtx.Watch(api.WatchOptions{}); err != nil {
		log.Fatalf("esbuild: failed to start JS watch mode: %s\n", err)
	}

	log.Println("esbuild: watching for asset changes... (Ctrl+C to stop)")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("esbuild: stopping watch mode...")
	cssCtx.Dispose()
	jsCtx.Dispose()
}

func copyStaticAssets() {
	if err := copyFile("node_modules/alpinejs/dist/cdn.min.js", "../static/alpinejs.min.js"); err != nil {
		log.Fatal(err)
	}
	if err := copyFile("node_modules/awesomplete/awesomplete.min.js", "../static/awesomplete.min.js"); err != nil {
		log.Fatal(err)
	}
	if err := copyFile("node_modules/htmx.org/dist/htmx.min.js", "../static/htmx.min.js"); err != nil {
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
	defer func() {
		if err := srcFile.Close(); err != nil {
			log.Fatalf("failed to close %s: %s", src, err)
		}
	}()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			log.Fatalf("failed to close %s: %s", dest, err)
		}
	}()

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
