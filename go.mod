module github.com/cytown/gintool

require (
	github.com/gin-contrib/logger v0.0.1
	github.com/gin-contrib/multitemplate v0.0.0-20190301062633-f9896279eead
	github.com/gin-gonic/gin v1.4.0
	github.com/go-errors/errors v1.0.1
	github.com/kr/pretty v0.1.0 // indirect
	github.com/rs/zerolog v1.9.1
	github.com/stretchr/testify v1.3.0
	github.com/v2pro/plz v0.0.0-20180227161703-2d49b86ea382
	golang.org/x/sys v0.0.0-20190516014833-cab07311ab81 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace (
	github.com/cytown/gintool => ../gin-tool

	github.com/gin-contrib/logger => /go/src/github.com/gin-contrib/logger
	github.com/gin-contrib/multitemplate => /go/src/github.com/gin-contrib/multitemplate
	github.com/ugorji/go => /go/src/github.com/ugorji/go

	golang.org/x/tools@v0.0.0-20190424220101-1e8e1cfdf96b => golang.org/x/tools v0.0.0-20190509153222-73554e0f7805
)
