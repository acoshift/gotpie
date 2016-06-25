package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/unrolled/render"
	"github.com/urfave/cli"
)

type writer struct {
	bytes.Buffer
}

func (writer) Header() http.Header {
	return http.Header{}
}

func (writer) WriteHeader(int) {
}

func main() {
	app := cli.NewApp()
	app.Name = "gotpie"
	app.Version = "0.0.1"
	app.Usage = "Compile your go template to html"
	app.ArgsUsage = "name"
	app.Commands = cli.Commands{
		cli.Command{
			Name:      "compile",
			ShortName: "c",
			Usage:     "Compile template to html",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "layout", Value: "layout"},
				cli.StringFlag{Name: "dir", Value: ""},
				cli.StringFlag{Name: "out", Value: "."},
				cli.BoolFlag{Name: "watch"},
			},
			Action: func(c *cli.Context) error {
				dir := c.String("dir")
				out := c.String("out")
				fn := c.Args().First()
				r := render.New(render.Options{
					Layout:                    c.String("layout"),
					Directory:                 dir,
					DisableHTTPErrorRendering: false,
					IsDevelopment:             true,
				})
				compile := func() error {
					b := writer{}
					if err := r.HTML(&b, 0, fn, nil); err != nil {
						return err
					}
					ioutil.WriteFile(out+"/"+fn+".html", b.Bytes(), 0644)
					return nil
				}

				if c.Bool("watch") {
					w, err := fsnotify.NewWatcher()
					defer w.Close()
					if err != nil {
						return err
					}
					w.Add(dir)
					compile()
					for {
						select {
						case ev := <-w.Events:
							if ev.Op != fsnotify.Chmod {
								compile()
							}
						case err := <-w.Errors:
							log.Println("error: ", err)
						}
					}
				} else {
					return compile()
				}
			},
		},
	}

	app.Run(os.Args)
}
