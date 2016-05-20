package main

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/codegangsta/cli"
)

func main() {
	var file string
	var bucket string
	i := impl{}

	app := cli.NewApp()
	app.Name = "bolter"
	app.Usage = "view boltdb file in your terminal"
	app.Version = "1.0.0"
	app.Author = "Hasit Mistry"
	app.Email = "hasit@uw.edu"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Usage:       "BoltDB file to view `FILE`",
			Destination: &file,
		},
		cli.StringFlag{
			Name:        "bucket, b",
			Usage:       "Bucket to view `BUCKET`",
			Destination: &bucket,
		},
	}
	app.Action = func(c *cli.Context) error {
		if file != "" {
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
		} else {
			cli.ShowAppHelp(c)
		}
		return nil
	}
	app.Run(os.Args)
}

type impl struct {
	DB *bolt.DB
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
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			items = append(items, item{Key: string(k), Value: string(v)})
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Key\tValue:\n")
	for _, item := range items {
		fmt.Printf("%s\t%s\n", item.Key, item.Value)
	}
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
	fmt.Printf("Buckets:\n")
	for _, b := range buckets {
		fmt.Printf("%s\n", b.Name)
	}
}
