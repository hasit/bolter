package main

import (
	"bufio"
	"fmt"
	"github.com/codegangsta/cli"
	kval "github.com/kval-access-language/kval-boltdb"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"strings"
)

var instructionLine = "> Enter bucket to explore (CTRL-Q to quit, CTRL-B to go back, ENTER to reset):\n\n"

func main() {
	var file string

	cli.AppHelpTemplate = `NAME:
  {{.Name}} - {{.Usage}}

VERSION:
  {{.Version}}

USAGE:
  {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}

GLOBAL OPTIONS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
AUTHOR:
  {{range .Authors}}{{ . }}{{end}}
COPYRIGHT:
  {{.Copyright}}
`
	app := cli.NewApp()
	app.Name = "bolter"
	app.Usage = "view boltdb file interactively in your terminal"
	app.Version = "2.0.0-kval-fork"
	app.Author = "Originally by Hasit Mistry. Interactive mode: Ross Spencer"
	app.Copyright = "(c) 2016 Hasit Mistry"
	//app.Email = ""
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Usage:       "boltdb `FILE` to view",
			Destination: &file,
		},
	}
	app.Action = func(c *cli.Context) error {
		if file == "" {
			cli.ShowAppHelp(c)
			return nil
		}

		var i impl
		i = impl{fmt: &tableFormatter{}}
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Fatal(err)
			return err
		}
		i.initDB(file)
		defer kval.Disconnect(i.kb)

		i.readInput()

		return nil
	}
	app.Run(os.Args)
}

func (i *impl) readInput() {
	i.listBuckets()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		bucket := scanner.Text()
		fmt.Fprintln(os.Stdout, "")
		switch bucket {
		case "\x11":
			return
		case "\x02":
			// TODO: Change KVAL to get first record...
			if !strings.Contains(i.loc, "") || !strings.Contains(i.loc, ">>") {
				fmt.Println("> Going back...")
				i.loc = ""
				i.listBuckets()
			} else {
				i.listBucketItems(bucket, true)
			}
		case "":
			i.listBuckets()
		default:
			i.listBucketItems(bucket, false)
		}
		bucket = ""
	}
}

type formatter interface {
	DumpBuckets([]bucket)
	DumpBucketItems(string, []item)
}

type impl struct {
	kb     kval.Kvalboltdb
	fmt    formatter
	bucket string
	loc    string // where we are in the structure
	cur    string
	root   bool // are we on root
}

type item struct {
	Key   string
	Value string
}

type bucket struct {
	Name string
}

func (i *impl) initDB(file string) {
	var err error
	// Connect to KVAL
	i.kb, err = kval.Connect(file)
	if err != nil {
		log.Fatal(err)
	}
}

func (i *impl) updateLoc(bucket string, goBack bool) string {

	// we've probably an invalid value and want to display
	// ourselves again...
	if bucket == i.cur {
		i.loc = bucket
		return i.loc
	}

	// handle goback
	if goBack {
		s := strings.Split(i.loc, ">>")
		i.loc = strings.Join(s[:len(s)-1], ">>")
		i.bucket = strings.Trim(s[len(s)-2], " ")
		return i.loc
	}

	// handle location on merit...
	if i.loc == "" {
		i.loc = bucket
		i.bucket = bucket
	} else {
		i.loc = i.loc + " >> " + bucket
		i.bucket = bucket
	}
	return i.loc
}

func (i *impl) listBucketItems(bucket string, goBack bool) {
	items := []item{}
	getItems := i.updateLoc(bucket, goBack)
	fmt.Fprintf(os.Stderr, "Query: "+getItems+"\n\n")
	if getItems != "" {
		res, err := kval.Query(i.kb, "GET "+getItems)
		if err != nil {
			if err.Error() != "Cannot GOTO bucket, bucket not found" {
				log.Fatal(err)
			} else {
				fmt.Fprintf(os.Stderr, "> Bucket not found\n")
				if i.root == true {
					i.listBuckets()
					return
				}
				i.listBucketItems(i.loc, true)
			}
		}
		if len(res.Result) == 0 {
			fmt.Fprintf(os.Stderr, "Invalid request.\n\n")
			i.listBucketItems(i.cur, false)
			return
		}

		for k, v := range res.Result {
			if v == kval.Nestedbucket {
				k = k + "*"
				v = ""
			}
			items = append(items, item{Key: string(k), Value: string(v)})
		}
		i.fmt.DumpBucketItems(i.bucket, items)
		i.root = false
		i.cur = getItems
		fmt.Fprint(os.Stdout, instructionLine)
	}
}

func (i *impl) listBuckets() {
	i.root = true
	i.loc = ""

	buckets := []bucket{}

	res, err := kval.Query(i.kb, "GET _")
	if err != nil {
		log.Fatal(err)
	}
	for k, _ := range res.Result {
		buckets = append(buckets, bucket{Name: string(k) + "*"})
	}

	fmt.Fprint(os.Stdout, "DB Layout:\n\n")
	i.fmt.DumpBuckets(buckets)
	fmt.Fprint(os.Stdout, instructionLine)
}

type tableFormatter struct{}

func (tf tableFormatter) DumpBuckets(buckets []bucket) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Buckets"})
	for _, b := range buckets {
		row := []string{b.Name}
		table.Append(row)
	}
	table.Render()
	fmt.Println()
}

func (tf tableFormatter) DumpBucketItems(bucket string, items []item) {
	fmt.Printf("Bucket: %s\n", bucket)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Value"})
	for _, item := range items {
		row := []string{item.Key, item.Value}
		table.Append(row)
	}
	table.Render()
	fmt.Println()
}
