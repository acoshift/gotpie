package main

import (
	"bytes"
	ejson "encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
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
				os.MkdirAll(out, 0755)
				r := render.New(render.Options{
					Layout:                    c.String("layout"),
					Directory:                 dir,
					Extensions:                []string{".tmpl"},
					DisableHTTPErrorRendering: false,
					IsDevelopment:             true,
					Funcs: []template.FuncMap{{
						"json": func(filename string) interface{} {
							bs, err := ioutil.ReadFile(filename)
							if err != nil {
								log.Println(err)
								return nil
							}
							var v interface{}
							err = ejson.Unmarshal(bs, &v)
							if err != nil {
								log.Println(err)
								return nil
							}
							return v
						}},
					},
				})
				m := minify.New()
				m.AddFunc("text/html", html.Minify)
				m.AddFunc("text/css", css.Minify)
				m.AddFunc("text/javascript", js.Minify)
				m.AddFunc("image/svg+xml", svg.Minify)
				m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
				m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
				compile := func() error {
					filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if info.IsDir() {
							return nil
						}

						path, _ = filepath.Rel(dir, path)
						fn := strings.TrimSuffix(path, filepath.Ext(path))
						if filepath.Ext(path) == ".tmpl" && filepath.Ext(fn) == ".entry" {
							fo := strings.TrimSuffix(fn, filepath.Ext(fn))
							log.Println("compile", path)
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
							ioutil.WriteFile(filepath.Join(out, fo+".html"), s, 0644)
						} else if filepath.Ext(path) != ".tmpl" && filepath.Base(path)[0] != '.' {
							log.Println("compile", path)
							b, _ := ioutil.ReadFile(filepath.Join(dir, path))
							if c.Bool("minify") {
								mm := mime.TypeByExtension(filepath.Ext(path))
								if mm != "" {
									s, err := m.Bytes(mm, b)
									if err == nil {
										b = s
									}
								}
							}
							o := filepath.Join(out, path)
							os.MkdirAll(filepath.Dir(o), 0755)
							ioutil.WriteFile(o, b, 0644)
						}
						return nil
					})
					return nil
				}

				if c.Bool("watch") {
					w, err := fsnotify.NewWatcher()
					if err != nil {
						log.Println(err)
						return err
					}
					defer w.Close()
					resetDir := func() {
						w.Close()
						w, _ = fsnotify.NewWatcher()
						w.Add(dir)
						filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if info.IsDir() {
								w.Add(path)
							}
							return nil
						})
					}
					resetDir()
					compile()
					for {
						select {
						case ev := <-w.Events:
							if ev.Op == fsnotify.Create || ev.Op == fsnotify.Remove || ev.Op == fsnotify.Rename {
								resetDir()
							}
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
