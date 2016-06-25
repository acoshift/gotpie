# gotpie

## Go Template Compiler

### Install

`go install github.com/acoshift/gotpie`

### Usage

`gotpie --watch --minify --out build src`

gotpie will compile all of .html files in src folder into build folder

### Example

Project structure

- src
  - footer.tmpl
  - layout.tmpl
  - menu.tmpl
  - index.html
  - about.html

`gotpie --layout layout --out build src`

Result

- build
  - index.html
  - about.html
