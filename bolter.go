package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	kval "github.com/kval-access-language/kval-boltdb"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

// Terminal lines...
const instructionLine = "> Enter bucket to explore (CTRL-X to quit, CTRL-B to go back, ENTER to go back to ROOT Bucket):"
const goingBack = "> Going back..."

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
	app.Version = "2.0.1"
	app.Authors = []*cli.Author{
		&cli.Author{
			Name:  "Hasit Mistry",
			Email: "hasitnm@gmail.com",
		},
	}
	app.Copyright = "(c) 2016 Hasit Mistry"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
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
		case "\x18":
			return
		case "\x02":
			if !strings.Contains(i.loc, "") || !strings.Contains(i.loc, ">>") {
				fmt.Fprintf(os.Stdout, "%s\n", goingBack)
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
	loc    string // navigation, what is our requested location in the store?
	cache  string // navigation, cache our last location to move back to
	root   bool   // navigation, are we @ root bucket?
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
	// Connect to KVAL using KVAL default mechanism
	// Can also use regular open plus perms, and kval.Attach()
	i.kb, err = kval.Connect(file)
	if err != nil {
		log.Fatal(err)
	}
}

func (i *impl) updateLoc(bucket string, goBack bool) string {

	// we've probably an invalid value and want to display
	// ourselves again...
	if bucket == i.cache {
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
	getQuery := i.updateLoc(bucket, goBack)
	if getQuery != "" {
		fmt.Fprintf(os.Stdout, "Query: "+getQuery+"\n\n")
		res, err := kval.Query(i.kb, "GET "+getQuery)
		if err != nil {
			if err.Error() != "Cannot GOTO bucket, bucket not found" {
				log.Fatal(err)
			} else {
				fmt.Fprintf(os.Stdout, "> Bucket not found\n")
				if i.root == true {
					i.listBuckets()
					return
				}
				i.listBucketItems(i.loc, true)
			}
		}
		if len(res.Result) == 0 {
			fmt.Fprintf(os.Stdout, "Invalid request.\n\n")
			i.listBucketItems(i.cache, false)
			return
		}

		for k, v := range res.Result {
			if v == kval.Nestedbucket {
				k = k + "*"
				v = ""
			}
			items = append(items, item{Key: string(k), Value: string(v)})
		}
		fmt.Fprintf(os.Stdout, "Bucket: %s\n", bucket)
		i.fmt.DumpBucketItems(i.bucket, items)
		i.root = false     // success this far means we're not at ROOT
		i.cache = getQuery // so we can also set the query cache for paging
		outputInstructionline()
	}
}

func (i *impl) listBuckets() {
	i.root = true
	i.loc = ""

	buckets := []bucket{}

	res, err := kval.Query(i.kb, "GET _") // KVAL: "GET _" will return ROOT
	if err != nil {
		log.Fatal(err)
	}
	for k := range res.Result {
		buckets = append(buckets, bucket{Name: string(k) + "*"})
	}

	fmt.Fprint(os.Stdout, "DB Layout:\n\n")
	i.fmt.DumpBuckets(buckets)
	outputInstructionline()
}

func outputInstructionline() {
	fmt.Fprintf(os.Stdout, "\n%s\n\n", instructionLine)
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
}

func (tf tableFormatter) DumpBucketItems(bucket string, items []item) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Value"})
	for _, item := range items {
		row := []string{item.Key, item.Value}
		table.Append(row)
	}
	table.Render()
}
