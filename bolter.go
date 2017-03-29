package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	kval "github.com/kval-access-language/kval-boltdb"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var file string
	var bucket string

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
`
	app := cli.NewApp()
	app.Name = "bolter"
	app.Usage = "view boltdb file in your terminal"
	app.Version = "1.0.0"
	app.Author = "Hasit Mistry"
	app.Email = "hasitnm@gmail.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Usage:       "boltdb `FILE` to view",
			Destination: &file,
		},
		cli.StringFlag{
			Name:        "bucket, b",
			Usage:       "boltdb `BUCKET` to view",
			Destination: &bucket,
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
		defer i.DB.Close()

		i.readInput()

		return nil
	}
	app.Run(os.Args)
}

func (i *impl) readInput() {
	fmt.Fprint(os.Stdout, "DB Layout:\n\n")
	i.listBuckets()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprint(os.Stdout, "\nEnter bucket to explore (q to quit, m to print):\n\n")
	for scanner.Scan() {
		bucket := scanner.Text()
		fmt.Fprintln(os.Stdout, "")
		switch(bucket) {
		case "q":
			fallthrough
		case "quit":
			return
		case "":
			i.listBuckets()
		default:
			i.listBucketItems(bucket)
		}
	}
}

type formatter interface {
	DumpBuckets([]bucket)
	DumpBucketItems(string, []item)
}

type impl struct {
	KV  kval.Kvalboltdb
	DB  *bolt.DB
	fmt formatter
	loc string		// where we are in the structure
}

type item struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type bucket struct {
	Name string `json:"Name"`
}

func (i *impl) initDB(file string) {
	var err error
	// Read-only permission, should be equiv. bolt.Open(file, 0400, nil)
	i.KV, err = kval.Connect(file)
	i.DB = kval.GetBolt(i.KV)
	if err != nil {
		log.Fatal(err)
	}
}

func (i *impl) updateLoc(bucket string) string {
	if i.loc == "" {
		i.loc = "GET " + bucket
	} else {
		i.loc = i.loc + " >> " + bucket
	}
	return i.loc
}

func (i *impl) listBucketItems(bucket string) {
	items := []item{}
	getItems := i.updateLoc(bucket)
	fmt.Fprintf(os.Stderr, "Query: " + i.loc + "\n\n")
	res, err := kval.Query(i.KV, getItems)
	if err != nil {
		if err.Error() != "Cannot GOTO bucket, bucket not found" {
			log.Fatal(err)
		} else {
			fmt.Fprintln(os.Stderr, "Bucket not found")
			fmt.Println(getItems)
		}
	}
	for k, v := range res.Result {
		if v == kval.Nestedbucket {
			k = k + "*"
			v = ""
		}
		items = append(items, item{Key: string(k), Value: string(v)})
	}

	i.fmt.DumpBucketItems(bucket, items)
}

func (i *impl) listBuckets() {
	buckets := []bucket{}
	err := i.DB.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(bucketname []byte, _ *bolt.Bucket) error {
			buckets = append(buckets, bucket{Name: string(bucketname)})
			return nil
		})
	})
	if err != nil {
		log.Fatal(err)
	}
	i.fmt.DumpBuckets(buckets)
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

