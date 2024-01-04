package main

import (
	"flag"
	"log"

	"go.callvis_core/analysis"
)

var (
	// focusFlag    = flag.String("focus", "main", "Focus specific package using name or import path.")
	groupFlag    = flag.String("group", "pkg", "Grouping functions by packages and/or types [pkg, type] (separated by comma)")
	limitFlag    = flag.String("limit", "", "Limit package paths to given prefixes (separated by comma)")
	ignoreFlag   = flag.String("ignore", "", "Ignore package paths containing given prefixes (separated by comma)")
	includeFlag  = flag.String("include", "", "Include package paths with given prefixes (separated by comma)")
	nostdFlag    = flag.Bool("nostd", false, "Omit calls to/from packages in standard library.")
	nointerFlag  = flag.Bool("nointer", false, "Omit calls to unexported functions.")
	testFlag     = flag.Bool("tests", false, "Include test code.")
	graphvizFlag = flag.Bool("graphviz", false, "Use Graphviz's dot program to render images.")
	httpFlag     = flag.String("http", ":7878", "HTTP service address.")
	skipBrowser  = flag.Bool("skipbrowser", false, "Skip opening browser.")
	outputFile   = flag.String("file", "", "output filename - omit to use server mode")
	outputFormat = flag.String("format", "svg", "output file format [svg | png | jpg | ...]")
	cacheDir     = flag.String("cacheDir", "", "Enable caching to avoid unnecessary re-rendering, you can force rendering by adding 'refresh=true' to the URL query or emptying the cache directory")
	debugFlag    = flag.Bool("debug", false, "Enable verbose log.")
	versionFlag  = flag.Bool("version", false, "Show version and exit.")
	transfer     = flag.String("transfer", "", "path|package|file ")
)

func main() {
	flag.Parse()
	args := flag.Args()
	log.Printf("transfer:%v", *transfer)
	log.Printf("args:%v\n", args)
	analy := analysis.DoAnalysis(args)
	callMap, err := analy.Render("")
	if err != nil {
		log.Printf("error:%v", err)
		return
	}
	analy.PrintOutput(callMap, *transfer)
}
