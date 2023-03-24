package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"

	"github.com/lifei6671/esdump/es"
	"github.com/lifei6671/esdump/esv7"
)

var version = "v0.1.0"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	app := cli.NewApp()
	config := es.Config{}
	app.Name = "esdump"
	app.Usage = "A CLI tool for exporting data from Elasticsearch into a file"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "query",
			Aliases:     []string{"q"},
			Usage:       "Query filename in Lucene syntax.",
			Destination: &config.Query,
		},
		&cli.MultiStringFlag{
			Target: &cli.StringSliceFlag{
				Name:    "match-all",
				Aliases: []string{"m"},
				Usage:   "List of <field>:<direction> pairs to filter.",
			},
			Destination: &config.MatchAll,
		},
		&cli.StringFlag{
			Name:        "output-file",
			Aliases:     []string{"o"},
			Usage:       "CSV file location. [required]",
			Required:    true,
			Destination: &config.Filename,
		},
		&cli.MultiStringFlag{
			Target: &cli.StringSliceFlag{
				Name:    "es-server",
				Aliases: []string{"e"},
				Usage:   "Elasticsearch host URL.",
			},
			Value:       []string{"http://localhost:9200"},
			Destination: &config.EsServer,
		},
		&cli.StringFlag{
			Name:        "auth",
			Aliases:     []string{"a"},
			Usage:       "Elasticsearch basic authentication in the form of username:password.",
			Destination: &config.Auth,
		},
		&cli.StringFlag{
			Name:        "es-version",
			Aliases:     []string{"E"},
			Usage:       "Elasticsearch version",
			Value:       "v7",
			Destination: &config.EsVersion,
		},
		&cli.MultiStringFlag{
			Target: &cli.StringSliceFlag{
				Name:    "index-prefixes",
				Aliases: []string{"i"},
				Usage:   "Index name prefix(es). Default is ['logstash-*'].",
			},
			Value:       []string{"log-*"},
			Destination: &config.Index,
		},
		&cli.MultiStringFlag{
			Target: &cli.StringSliceFlag{
				Name:     "fields",
				Aliases:  []string{"f"},
				Usage:    "List of selected fields in output. ",
				Required: true,
			},
			Destination: &config.Fields,
		},
		&cli.MultiStringFlag{
			Target: &cli.StringSliceFlag{
				Name:    "sort",
				Aliases: []string{"s"},
				Usage:   "List of <field>:<desc|asc> pairs to sort on.",
			},
			Destination: &config.Sort,
		},
		&cli.IntFlag{
			Name:        "page-size",
			Aliases:     []string{"p"},
			Usage:       "Maximum number returned per page.",
			Value:       1000,
			Destination: &config.MaxSize,
		},
		&cli.DurationFlag{
			Name:        "scroll-size",
			Aliases:     []string{"S"},
			Usage:       "Scroll size for each batch of results. ",
			Value:       time.Minute * 5,
			Destination: &config.Scroll,
		},
		&cli.StringFlag{
			Name:        "range-field",
			Aliases:     []string{"R"},
			Usage:       "scope field for query",
			Value:       "@timestamp",
			Destination: &config.RangeField,
		},
		&cli.MultiStringFlag{
			Target: &cli.StringSliceFlag{
				Name:    "range-value",
				Aliases: []string{"V"},
				Usage:   "List of <field>:<direction> pairs to range on.",
			},
			Value:       []string{time.Now().Add(-24 * time.Hour).Format(time.RFC3339Nano), time.Now().Format(time.RFC3339Nano)},
			Destination: &config.RangeValue,
		},
		&cli.StringFlag{
			Name:        "raw-query",
			Aliases:     []string{"r"},
			Usage:       "Switch query format in the Query DSL.",
			Destination: &config.RawQuery,
		},
		&cli.BoolFlag{
			Name:        "ignore-err",
			Aliases:     []string{"n"},
			Usage:       "Ignore non-fatal error messages.",
			Value:       true,
			Destination: &config.IgnoreErr,
		},
		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "Debug mode on.",
			Value:       true,
			Destination: &config.Debug,
		},
	}
	app.Version = version
	app.Action = func(ctx *cli.Context) error {
		if config.Debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
		log.Debug().Str("es_version", config.EsVersion).
			Strs("fields", config.Fields).
			Send()

		var client es.Client
		var err error
		if config.EsVersion == "v7" {
			client, err = esv7.NewClient(config)
		}
		if err != nil {
			log.Fatal().Err(err)
			return err
		}
		if client == nil {
			return fmt.Errorf("elasticsearch version does not support:%v", config.EsVersion)
		}
		receive := make(chan json.RawMessage, 1)
		go func() {
			err = client.Dump(ctx.Context, receive)
			if err != nil {
				log.Fatal().Err(err)
			}
		}()

		f, err := os.Create(config.Filename)
		if err != nil {
			log.Fatal().Err(err)
		}
		defer f.Close()
		writer := es.NewCSVWriter(f, config.Fields)
		defer writer.Close()

		//for doc := range receive {
		//	writer.Write(context.Background(), doc)
		//}
		doc := <-receive
		log.Debug().Msg(string(doc))
		writer.Write(context.Background(), doc)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Send()
	}
}
