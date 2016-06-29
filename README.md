# gotpie

## Go Template Compiler

### Install

`go install github.com/acoshift/gotpie`

### Usage

`gotpie compile --watch --out build src`

gotpie will compile all of .entry.tmpl files in src folder into build folder

### Example

Project structure

- src
  - footer.tmpl
  - layout.tmpl
  - menu.tmpl
  - index.entry.tmpl
  - about.entry.tmpl

`gotpie compile --layout layout --out build src`

Result

- build
  - index.html
  - about.html
