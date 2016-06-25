package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
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
	app.ArgsUsage = "dir"
	app.Commands = cli.Commands{
		cli.Command{
			Name:      "compile",
			ShortName: "c",
			Usage:     "Compile template to html",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "layout", Value: "layout"},
				cli.StringFlag{Name: "out", Value: "."},
				cli.BoolFlag{Name: "watch"},
				cli.BoolFlag{Name: "minify"},
			},
			Action: func(c *cli.Context) error {
				dir := c.Args().First()
				out := c.String("out")
				r := render.New(render.Options{
					Layout:                    c.String("layout"),
					Directory:                 dir,
					Extensions:                []string{".tmpl", ".html"},
					DisableHTTPErrorRendering: false,
					IsDevelopment:             true,
				})
				m := minify.New()
				m.Add("text/html", &html.Minifier{KeepDefaultAttrVals: true})
				compile := func() error {
					var err error
					fns, _ := filepath.Glob(filepath.Join(dir, "*.html"))
					for _, fn := range fns {
						fn, _ = filepath.Rel(dir, fn)
						fn = strings.TrimSuffix(fn, filepath.Ext(fn))
						log.Println("compile " + fn)
						b := writer{}
						if err = r.HTML(&b, 0, fn, nil); err != nil {
							log.Println(err)
							return err
						}
						s := b.Bytes()
						if c.Bool("minify") {
							s, err = m.Bytes("text/html", s)
							if err != nil {
								s = b.Bytes()
							}
						}
						ioutil.WriteFile(filepath.Join(out, fn+".html"), s, 0644)
					}
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
