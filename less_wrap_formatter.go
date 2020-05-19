package main

import (
	"io"
	"os/exec"
)

type moreWrapFormatter struct {
	formatter formatter
}

func (mf moreWrapFormatter) wrapDump(w io.Writer, dump func(io.Writer)) {
	lessCmd := exec.Command("more")
	pipeR, pipeW := io.Pipe()
	go func() {
		dump(pipeW)
		pipeW.Close()
	}()
	lessCmd.Stdin = pipeR
	lessCmd.Stdout = w
	lessCmd.Run()
}

func (mf moreWrapFormatter) DumpBuckets(w io.Writer, buckets []bucket) {
	mf.wrapDump(w, func(w io.Writer) {
		mf.formatter.DumpBuckets(w, buckets)
	})
}

func (mf moreWrapFormatter) DumpBucketItems(w io.Writer, bucket string, items []item) {
	mf.wrapDump(w, func(w io.Writer) {
		mf.formatter.DumpBucketItems(w, bucket, items)
	})
}
