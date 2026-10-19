// Command experiment-report summarises results CSV files written by cmd/experiment into the final
// tables of the experiment: detection latency per detector, stream vs batch,
// false positive rate in baseline periods.
//
// Usage (from the repository root):
//
//	go run ./cmd/experiment-report experiment-results/*.csv
//	go run ./cmd/experiment-report -csv summary.csv experiment-results/payment_outage-*.csv
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/squ1ky/flyte/internal/experiment"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "experiment-report:", err)
		os.Exit(1)
	}
}

func run() error {
	csvOut := flag.String("csv", "", "also write the latency table as CSV to this file")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "usage: experiment-report [-csv out.csv] results.csv...")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		return fmt.Errorf("no results files given")
	}

	rows, err := experiment.ReadResults(flag.Args()...)
	if err != nil {
		return err
	}
	report := experiment.BuildReport(rows)

	if err := report.WriteMarkdown(os.Stdout); err != nil {
		return err
	}
	if *csvOut == "" {
		return nil
	}
	f, err := os.Create(*csvOut)
	if err != nil {
		return err
	}
	if err := report.WriteLatencyCSV(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
