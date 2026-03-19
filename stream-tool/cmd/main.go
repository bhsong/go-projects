package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bhsong/go-projects/stream-tool/internal/stream"
)

func main() {
	teeFile := flag.String("tee", "", "...")
	outFile := flag.String("out", "", "...")
	flag.Parse()

	if len(flag.Args()) == 0 {
		printUsage(os.Stdout)
		os.Exit(0)
	}

	var r io.Reader
	path := flag.Args()[0]
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "오류: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	r = f

	if *teeFile != "" {
		tf, err := os.Create(*teeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stderr: 오류: %v\n", err)
			os.Exit(1)
		}
		defer tf.Close()
		r = stream.Tee(r, tf)
	}

	var w io.Writer
	if *outFile != "" {
		of, err := os.Create(*outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stderr: 오류: %v\n", err)
			os.Exit(1)
		}
		defer of.Close()
		w = stream.MultiWrite(os.Stdout, of)
	} else {
		w = os.Stdout
	}

	stats, err := stream.Count(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stderr: 오류: %v\n", err)
		os.Exit(1)
	}
	printStats(w, stats)

}

func printStats(w io.Writer, s stream.Stats) {
	fmt.Fprintf(w, "줄: %d\n단어: %d\n바이트: %d\n", s.Lines, s.Words, s.Bytes)
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "사용법: stream-tool [--tee <file>] [--out <file>] [<input-file>]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "인자:")
	fmt.Fprintln(w, "  <input-file>   읽을 파일 (생략 시 stdin)")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "플래그:")
	fmt.Fprintln(w, "  --tee <file>   원본 내용을 파일에 동시 저장")
	fmt.Fprintln(w, "  --out <file>   결과(줄/단어/바이트)를 파일에 동시 저장")
}
