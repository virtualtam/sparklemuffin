// Copyright VirtualTam 2022, 2026
// SPDX-License-Identifier: MIT

// Package main provides an esbuild pipeline to minify and bundle static assets used by
// the SparkleMuffin Web application.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/bep/godartsass/v2"
	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	watchMode := flag.Bool("watch", false, "Watch for changes and rebuild automatically")
	flag.Parse()

	copyStaticAssets()
	generateChromaCss()
	compileSass()

	if *watchMode {
		watchAssets()
	} else {
		buildAssets()
	}
}

// copyStaticAssets copies static assets as-is.
func copyStaticAssets() {
	if err := copyFile("node_modules/alpinejs/dist/cdn.min.js", "../static/alpinejs.min.js"); err != nil {
		log.Fatal(err)
	}
	if err := copyFile("node_modules/htmx.org/dist/htmx.min.js", "../static/htmx.min.js"); err != nil {
		log.Fatal(err)
	}
	if err := copyFile("node_modules/bootstrap/dist/js/bootstrap.bundle.min.js", "../static/bootstrap.bundle.min.js"); err != nil {
		log.Fatal(err)
	}
	if err := copyFiles("favicons", "../static"); err != nil {
		log.Fatal(err)
	}
}

// generateChromaCss generates the CSS for Markdown code block syntax
// highlighting: "nord-light" unscoped (the default), "nord" nested under
// [data-bs-theme="dark"] so it follows the runtime theme toggle. Goldmark
// itself (internal/http/www/view/markdown.go) only emits token classes, not
// colours, so this is the sole place the actual palette is defined.
//
// The generated CSS is imported in the main CSS file (www.css) and processed as part of the assets pipeline.
//
// This allows:
// - configuring Chroma to output CSS classes instead of inline style;
// - enforcing a strict Content Security Policy by not having to allow 'unsafe-inline' styles.
func generateChromaCss() {
	registerNordLightStyle()

	formatter := html.New(html.WithClasses(true))

	var buf bytes.Buffer
	if err := formatter.WriteCSS(&buf, styles.Get("nord-light")); err != nil {
		log.Fatalf("esbuild: failed to generate light chroma CSS: %s\n", err)
	}

	var darkBuf bytes.Buffer
	if err := formatter.WriteCSS(&darkBuf, styles.Get("nord")); err != nil {
		log.Fatalf("esbuild: failed to generate dark chroma CSS: %s\n", err)
	}
	buf.WriteString("[data-bs-theme=\"dark\"] {\n")
	buf.Write(darkBuf.Bytes())
	buf.WriteString("}\n")

	if err := writeFile(&buf, "css/chroma.css"); err != nil {
		log.Fatalf("esbuild: failed to write chroma CSS: %s\n", err)
	}
}

// nordLightEntries mirrors chroma's own bundled "nord" style
// (github.com/alecthomas/chroma/v2/styles/nord.xml), with the background and
// base text swapped for light-mode contrast and every accent colour shaded
// to clear a 4.5:1 contrast ratio against the light body background.
var nordLightEntries = chroma.StyleEntries{
	chroma.Background:            "#2e3440 bg:#e5e9f0",
	chroma.Error:                 "#994e55",
	chroma.Keyword:               "bold #4d6174",
	chroma.KeywordPseudo:         "nobold #4d6174",
	chroma.KeywordType:           "nobold #4d6174",
	chroma.Name:                  "#2e3440",
	chroma.NameAttribute:         "#4f6767",
	chroma.NameBuiltin:           "#4d6174",
	chroma.NameClass:             "#4f6767",
	chroma.NameConstant:          "#4f6767",
	chroma.NameDecorator:         "#875849",
	chroma.NameEntity:            "#875849",
	chroma.NameException:         "#994e55",
	chroma.NameFunction:          "#446068",
	chroma.NameLabel:             "#4f6767",
	chroma.NameNamespace:         "#4f6767",
	chroma.NameOther:             "#2e3440",
	chroma.NameTag:               "#4d6174",
	chroma.NameVariable:          "#2e3440",
	chroma.NameProperty:          "#4f6767",
	chroma.LiteralString:         "#5a694d",
	chroma.LiteralStringDoc:      "#525e73",
	chroma.LiteralStringEscape:   "#5e5138",
	chroma.LiteralStringInterpol: "#5a694d",
	chroma.LiteralStringOther:    "#5a694d",
	chroma.LiteralStringRegex:    "#5e5138",
	chroma.LiteralStringSymbol:   "#5a694d",
	chroma.LiteralNumber:         "#755c70",
	chroma.Operator:              "#4d6174",
	chroma.OperatorWord:          "bold #4d6174",
	chroma.Punctuation:           "#2e3440",
	chroma.Comment:               "italic #525e73",
	chroma.CommentPreproc:        "#4b678a",
	chroma.GenericDeleted:        "#994e55",
	chroma.GenericEmph:           "italic",
	chroma.GenericError:          "#994e55",
	chroma.GenericHeading:        "bold #446068",
	chroma.GenericInserted:       "#5a694d",
	chroma.GenericOutput:         "#2e3440",
	chroma.GenericPrompt:         "bold #4c566a",
	chroma.GenericStrong:         "bold",
	chroma.GenericSubheading:     "bold #446068",
	chroma.GenericTraceback:      "#994e55",
	chroma.TextWhitespace:        "#2e3440",
}

// registerNordLightStyle builds "nord-light" and pairs it with chroma's bundled "nord" as its dark counterpart.
func registerNordLightStyle() {
	nordLight, err := chroma.NewStyleBuilder("nord-light").AddAll(nordLightEntries).Build()
	if err != nil {
		log.Fatalf("chroma: failed to build the nord-light style: %s\n", err)
	}

	styles.RegisterPair(nordLight, styles.Get("nord"))
}

// sassEntryPoints lists the custom Sass entry points compiled ahead of the
// esbuild CSS bundling pass. See scss/bootstrap.scss.
var sassEntryPoints = map[string]string{
	"scss/bootstrap.scss": "css/bootstrap.css",
}

// compileSass compiles the custom Sass entry points to plain CSS using Dart
// Sass, ahead of the esbuild CSS bundling pass.
func compileSass() {
	transpiler, err := godartsass.Start(godartsass.Options{
		DartSassEmbeddedFilename: sassEmbeddedBinaryPath(),
	})
	if err != nil {
		log.Fatalf("sass: failed to start the Dart Sass compiler: %s\n", err)
	}

	for src, dst := range sassEntryPoints {
		if err := compileSassFile(transpiler, src, dst); err != nil {
			_ = transpiler.Close()
			log.Fatalf("sass: %s\n", err)
		}
	}

	if err := transpiler.Close(); err != nil {
		log.Fatalf("sass: failed to close the Dart Sass compiler: %s\n", err)
	}
}

// compileSassFile compiles one Sass entry point to plain CSS.
func compileSassFile(transpiler *godartsass.Transpiler, src, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", src, err)
	}

	result, err := transpiler.Execute(godartsass.Args{
		Source: mustReadFile(src),
		URL:    fileURL(absSrc),
	})
	if err != nil {
		return fmt.Errorf("failed to compile %s: %w", src, err)
	}

	if err := writeFile(strings.NewReader(result.CSS), dst); err != nil {
		return fmt.Errorf("failed to write %s: %w", dst, err)
	}

	return nil
}

// sassEmbeddedBinaryPath returns the path to the platform-specific Dart Sass
// embedded compiler binary installed as an npm optional dependency of
// "sass-embedded".
func sassEmbeddedBinaryPath() string {
	npmArch := runtime.GOARCH
	if npmArch == "amd64" {
		npmArch = "x64"
	}

	npmPlatform := runtime.GOOS
	binName := "sass"
	if npmPlatform == "windows" {
		npmPlatform = "win32"
		binName = "sass.bat"
	}

	return filepath.Join(
		"node_modules",
		fmt.Sprintf("sass-embedded-%s-%s", npmPlatform, npmArch),
		"dart-sass",
		binName,
	)
}

// fileURL converts an absolute filesystem path to a "file://" URL, as
// expected by godartsass to resolve relative Sass imports.
func fileURL(absPath string) string {
	path := filepath.ToSlash(absPath)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	u := url.URL{Scheme: "file", Path: path}
	return u.String()
}

// mustReadFile reads the contents of path, or exits the program on error.
func mustReadFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("sass: failed to read %s: %s\n", path, err)
	}
	return string(data)
}

var (
	// cssBuildOptions configures how CSS files are processed by esbuild.
	cssBuildOptions = api.BuildOptions{
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

	// jsBuildOptions configure how JavaScript files are processed by esbuild.
	jsBuildOptions = api.BuildOptions{
		EntryPoints: []string{
			"js/bootstrap-modal-bridge.js",
			"js/complete-tags.js",
			"js/easymde-init.js",
			"js/theme-toggle.js",
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
)

// buildAssets processes CSS and JavaScript assets once.
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

// watchAssets watches for changes in CSS and JavaScript assets and processes them when necessary.
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

// writeFile creates a file and its parent directories, and writes the contents of r to it.
func writeFile(r io.Reader, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close %s: %s", path, err)
		}
	}(f)

	if _, err := io.Copy(f, r); err != nil {
		return err
	}

	log.Println("esbuild: wrote", path)
	return nil
}

// copyFile creates the destination file and its parent directories, and copies the contents of the source file to it.
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

// copyFiles creates the destination directory and recursively copies the contents of the source directory to it.
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
