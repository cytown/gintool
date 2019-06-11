module github.com/cytown/gintool

require (
	github.com/FlowerWrong/plushgin v0.0.0-20171223094000-bad75070f0c8
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gin-contrib/logger v0.0.1
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/go-errors/errors v1.0.1
	github.com/gobuffalo/helpers v0.0.0-20190506214229-8e6f634af7c3 // indirect
	github.com/gobuffalo/plush v3.8.2+incompatible // indirect
	github.com/gobuffalo/uuid v2.0.5+incompatible // indirect
	github.com/gobuffalo/validate v2.0.3+incompatible // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/rs/zerolog v1.14.3
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/ugorji/go v1.1.5-pre // indirect
	github.com/v2pro/plz v0.0.0-20180227161703-2d49b86ea382
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65 // indirect
	golang.org/x/sys v0.0.0-20190602015325-4c4f7f33c9ed // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace (
	github.com/cytown/gintool => ../gintool

	golang.org/x/tools@v0.0.0-20190424220101-1e8e1cfdf96b => golang.org/x/tools v0.0.0-20190509153222-73554e0f7805
)
