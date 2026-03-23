package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bhsong/go-projects/stream-tool/internal/crypto"
	"github.com/bhsong/go-projects/stream-tool/internal/stream"
)

func main() {
	if len(os.Args) < 2 {
		printUsage(os.Stdout)
		return
	}

	var err error

	switch os.Args[1] {
	case "count":
		err = runCount(os.Args[2:])
	case "hash":
		err = runHash(os.Args[2:])
	case "verify":
		err = runVerify(os.Args[2:])
	case "hmac":
		err = runHMAC(os.Args[2:])
	case "hmac-verify":
		err = runHMACVerify(os.Args[2:])
	case "encrypt":
		err = runEncrypt(os.Args[2:])
	case "decrypt":
		err = runDecrypt(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runCount(args []string) error {
	fs := flag.NewFlagSet("count", flag.ContinueOnError)
	teeFile := fs.String("tee", "", "copy input to file while counting")
	outFile := fs.String("out", "", "write stats to file in addition to stdout")

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("runCount: %w", err)
	}

	var r io.Reader
	if fs.NArg() == 0 {
		r = os.Stdin
	} else {
		path := fs.Arg(0)
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("runCount: %w", err)
		}
		defer f.Close()

		r = f
	}
	if *teeFile != "" {
		tf, err := os.Create(*teeFile)
		if err != nil {
			return fmt.Errorf("runCount: create tee file: %w", err)
		}
		defer tf.Close()

		r = stream.Tee(r, tf)
	}
	var w io.Writer = os.Stdout
	if *outFile != "" {
		of, err := os.Create(*outFile)
		if err != nil {
			return fmt.Errorf("runCount: create out file: %w", err)
		}
		defer of.Close()

		w = stream.MultiWrite(os.Stdout, of)
	}

	stats, err := stream.Count(r)
	if err != nil {
		return fmt.Errorf("runCount: %w", err)
	}

	fmt.Fprintf(w, "줄: %d\n단어: %d\n바이트: %d\n", stats.Lines, stats.Words, stats.Bytes)

	return nil
}

func runHash(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: stream-tool hash <file>")
	}

	path := args[0]

	hash, err := crypto.HashFile(path)
	if err != nil {
		return fmt.Errorf("runHash: %w", err)
	}

	fmt.Println(hash)

	return nil
}

func runVerify(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: stream-tool verify <file> <hash>")
	}

	path, expected := args[0], args[1]

	ok, err := crypto.VerifyFile(path, expected)
	if err != nil {
		return fmt.Errorf("runVerify: %w", err)
	}

	if ok {
		fmt.Println("OK: hash matches")
	} else {
		return crypto.ErrHashMistmatch
	}

	return nil
}

func runHMAC(args []string) error {
	fs := flag.NewFlagSet("hmac", flag.ContinueOnError)

	key := fs.String("key", "", "secret key")

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("runHMAC: %w", err)
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: stream-tool hmac --key <secret> <file>")
	}

	if *key == "" {
		return fmt.Errorf("runHMAC: --key is required")
	}

	path := fs.Arg(0)

	mac, err := crypto.GenerateHMAC(path, *key)
	if err != nil {
		return fmt.Errorf("runHMAC: %w", err)
	}

	fmt.Println(mac)

	return nil
}

func runHMACVerify(args []string) error {
	fs := flag.NewFlagSet("hmac-verify", flag.ContinueOnError)
	key := fs.String("key", "", "secret key")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("runHMACVerify: %w", err)
	}

	if fs.NArg() < 2 {
		return fmt.Errorf("usage: stream-tool hmac-verify --key <secret> <file> <mac>")
	}

	if *key == "" {
		return fmt.Errorf("runHMACVerify: --key is required")
	}

	path, expected := fs.Arg(0), fs.Arg(1)
	ok, err := crypto.VerifyHMAC(path, *key, expected)
	if err != nil {
		return fmt.Errorf("runHMACVerify: %w", err)
	}

	if ok {
		fmt.Println("OK: hmac matches")
	} else {
		return crypto.ErrHMACMistmatch
	}
	return nil
}

func runEncrypt(args []string) error {
	fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	pass := fs.String("pass", "", "encryption password")
	out := fs.String("out", "", "output file path")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("runEncrypt: %w", err)
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: stream-tool encrypt --pass <password> --out <file> <file>")
	}

	if *pass == "" {
		return fmt.Errorf("runEncrypt: --pass is required")
	}

	if *out == "" {
		return fmt.Errorf("runEncrypt: --out is required")
	}

	src := fs.Arg(0)
	err := crypto.EncryptFile(src, *out, *pass)
	if err != nil {
		return fmt.Errorf("runEncrypt: %w", err)
	}

	fmt.Printf("encrypted: %s -> %s\n", src, *out)
	return nil
}

func runDecrypt(args []string) error {
	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	pass := fs.String("pass", "", "decryption password")
	out := fs.String("out", "", "output file path")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("runDecrypt: %w", err)
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: stream-tool decrypt --pass <password> --out <file> <file>")
	}

	if *pass == "" {
		return fmt.Errorf("runDecrypt: --pass is required")
	}

	if *out == "" {
		return fmt.Errorf("runDecrypt: --out is required")
	}

	src := fs.Arg(0)
	err := crypto.DecryptFile(src, *out, *pass)
	if err != nil {
		return fmt.Errorf("runDecrypt: %w", err)
	}

	fmt.Printf("decrypted: %s -> %s\n", src, *out)

	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: stream-tool <command> [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Stream commands:")
	fmt.Fprintln(w, "  count [file]             count lines, words, bytes (stdin if no file)")
	fmt.Fprintln(w, "    --tee <file>           copy input to file while counting")
	fmt.Fprintln(w, "    --out <file>           write stats to file in addition to stdout")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Crypto commands:")
	fmt.Fprintln(w, "  hash <file>              print SHA-256 hash of file")
	fmt.Fprintln(w, "  verify <file> <hash>     verify file matches expected SHA-256 hash")
	fmt.Fprintln(w, "  hmac <file>              generate HMAC-SHA256 of file")
	fmt.Fprintln(w, "    --key <secret>         secret key (required)")
	fmt.Fprintln(w, "  hmac-verify <file> <mac> verify HMAC-SHA256 of file")
	fmt.Fprintln(w, "    --key <secret>         secret key (required)")
	fmt.Fprintln(w, "  encrypt <file>           encrypt file with AES-256-GCM")
	fmt.Fprintln(w, "    --pass <password>      encryption password (required)")
	fmt.Fprintln(w, "    --out <file>           output file path (required)")
	fmt.Fprintln(w, "  decrypt <file>           decrypt file with AES-256-GCM")
	fmt.Fprintln(w, "    --pass <password>      decryption password (required)")
	fmt.Fprintln(w, "    --out <file>           output file path (required)")
}
