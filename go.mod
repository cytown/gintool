module github.com/cytown/gintool

require (
	github.com/gin-contrib/logger v0.0.1
	github.com/gin-contrib/multitemplate v0.0.0-20190528082104-30e424939505
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/go-errors/errors v1.0.1
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/rs/zerolog v1.14.3
	github.com/stretchr/testify v1.3.0
	github.com/ugorji/go v1.1.5-pre // indirect
	github.com/v2pro/plz v0.0.0-20180227161703-2d49b86ea382
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65 // indirect
	golang.org/x/sys v0.0.0-20190602015325-4c4f7f33c9ed // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace (
	github.com/cytown/gintool => ../gintool

	golang.org/x/tools@v0.0.0-20190424220101-1e8e1cfdf96b => golang.org/x/tools v0.0.0-20190509153222-73554e0f7805
)
