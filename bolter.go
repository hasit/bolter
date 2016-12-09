package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
	"github.com/olekukonko/tablewriter"
)

func main() {
	var file string
	var bucket string
	var machineFriendly bool

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
		cli.BoolFlag{
			Name:        "machine-friendly, m",
			Usage:       "key=value format",
			Destination: &machineFriendly,
		},
	}
	app.Action = func(c *cli.Context) error {
		if file == "" {
			cli.ShowAppHelp(c)
			return nil
		}

		var i impl
		if machineFriendly {
			i = impl{fmt: &machineFormatter{}}
		} else {
			i = impl{fmt: &tableFormatter{}}
		}
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Fatal(err)
			return err
		}
		i.initDB(file)
		defer i.DB.Close()
		if bucket != "" {
			i.listBucketItems(bucket)
		} else {
			i.listBuckets()
		}
		return nil
	}
	app.Run(os.Args)
}

type formatter interface {
	DumpBuckets([]bucket)
	DumpBucketItems(string, []item)
}

type impl struct {
	DB  *bolt.DB
	fmt formatter
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
	// Read-only permission
	i.DB, err = bolt.Open(file, 0400, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (i *impl) listBucketItems(bucket string) {
	items := []item{}
	err := i.DB.View(func(tx *bolt.Tx) error {
		if i := len(bucket) - 1; bucket[i:] == "." {
			bucket = bucket[:i]
		}
		nbs := strings.Split(bucket, ".")
		b := tx.Bucket([]byte(nbs[0]))
		if b == nil {
			return nil
		}
		for _, nb := range nbs[1:] {
			b = b.Bucket([]byte(nb))
			if b == nil {
				return nil
			}
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				k = append(k, byte('*'))
			}
			items = append(items, item{Key: string(k), Value: string(v)})
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
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
}

type machineFormatter struct{}

func (mf machineFormatter) DumpBuckets(buckets []bucket) {
	for _, b := range buckets {
		fmt.Println(b.Name)
	}
}

func (mf machineFormatter) DumpBucketItems(_ string, items []item) {
	for _, item := range items {
		fmt.Printf("%s=%s\n", item.Key, item.Value)
	}
}
