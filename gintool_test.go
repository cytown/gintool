package gintool

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewGin(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *GinEngine
		wantErr bool
	}{
		{
			name: "empty file",
			args: args{
				path: "",
			},
			want: &GinEngine{
				address: ":8080",
				statics: map[string]string{
					"/html":   "static",
					"/images": "static/images",
				},
				errors: map[int]string{
					404: "error/404.html",
					500: "error/500.html",
				},
				logfile:  "log/gin.log",
				errorlog: "log/gin.log",
			},
			wantErr: false,
		},
		{
			name: "testdata gin.conf file",
			args: args{
				path: "testdata/gin.conf",
			},
			want: &GinEngine{
				address: "localhost:8088",
				statics: map[string]string{
					"/html":   "testdata/static",
					"/images": "testdata/static/images",
				},
				staticFs: map[string]string{
					"/favicon.ico": "testdata/static/images/favicon.png",
				},
				errors: map[int]string{
					404: "testdata/error/404.html",
					500: "testdata/error/500.html",
				},
				logfile:  "/tmp/gin.log",
				errorlog: "/tmp/gin_error.log",
				keyFile:  "testdata/keyfile",
				certFile: "testdata/certfile",
			},
			wantErr: false,
		},
		{
			name: "not exist gin.conf file",
			args: args{
				path: "test/gin.conf",
			},
			want:    &GinEngine{},
			wantErr: true,
		},
		{
			name: "not gin.conf file",
			args: args{
				path: "testdata",
			},
			want:    &GinEngine{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGin(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGin() errors = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				return
			}
			got.Engine = tt.want.Engine
			got.template = tt.want.template
			//fmt.Println(tt.want.statics, got.statics)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGin() = \n%v\n, want \n%v", got, tt.want)
			}
		})
	}
}

func TestGinEngine_Start(t *testing.T) {
	type resp struct {
		code int
		body string
	}
	type fields struct {
		Engine   *gin.Engine
		address  string
		statics  map[string]string
		staticFs map[string]string
		error    map[int]string
		logfile  string
		errorlog string
		certFile string
		keyFile  string
		url      []string
	}
	tests := []struct {
		name    string
		fields  fields
		init    func(g *GinEngine)
		wantErr bool
		wantRes []resp
	}{
		{
			"wrong address",
			fields{
				address: "sdfasf",
				Engine:  gin.Default(),
			},
			nil,
			true,
			nil,
		},
		{
			"correct address",
			fields{
				address: "localhost:8080",
				Engine:  gin.Default(),
			},
			nil,
			false,
			nil,
		},
		{
			"correct address with static file",
			fields{
				address: "localhost:8080",
				Engine:  gin.Default(),
				statics: map[string]string{
					"/t": "testdata/static/",
				},
				staticFs: map[string]string{
					"/a.txt": "testdata/static/test.txt",
				},
				url: []string{
					"http://localhost:8080/a.txt",
					"http://localhost:8080/t/test.txt",
				},
			},
			nil,
			false,
			[]resp{
				resp{
					200,
					"hello world",
				},
				resp{
					200,
					"hello world",
				},
			},
		},
		{
			"test with not found",
			fields{
				address: "localhost:8080",
				Engine:  gin.Default(),
				url: []string{
					"http://localhost:8080/a.txt",
				},
			},
			nil,
			false,
			[]resp{
				resp{
					404,
					"404 page not found",
				},
			},
		},
		{
			"test with server error",
			fields{
				address: "localhost:8080",
				Engine:  gin.Default(),
				url: []string{
					"http://localhost:8080/",
				},
			},
			func(g *GinEngine) {
				g.Engine.Use(ginRecovery(map[int]string{}))
				g.Engine.GET("/", func(c *gin.Context) {
					panic("test only")
				})
			},
			false,
			[]resp{
				resp{
					500,
					"",
				},
			},
		},
	}
	resetDefault()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GinEngine{
				Engine:   tt.fields.Engine,
				address:  tt.fields.address,
				statics:  tt.fields.statics,
				staticFs: tt.fields.staticFs,
				errors:   tt.fields.error,
				logfile:  tt.fields.logfile,
				errorlog: tt.fields.errorlog,
				certFile: tt.fields.certFile,
				keyFile:  tt.fields.keyFile,
			}
			t1 := time.Now()
			go func() {
				if tt.init != nil {
					tt.init(g)
				}
				err := g.Start()
				log.Println("start return", err)
				if (err != nil) != tt.wantErr && !time.Now().After(t1.Add(20 * time.Millisecond)) {
					t.Errorf("GinEngine.Start() errors = %v, wantErr %v", err, tt.wantErr)
				}
			}()

			time.Sleep(10 * time.Millisecond)
			log.Println("started: ", g.server)
			for idx, url := range tt.fields.url {
				if url != "" {
					//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
					tr := &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					}
					client := &http.Client{Transport: tr}
					//_, err := client.Get("https://golang.org/")
					res, _ := client.Get(url)
					resp, _ := ioutil.ReadAll(res.Body)
					assert.Equal(t, tt.wantRes[idx].code, res.StatusCode)
					assert.Equal(t, tt.wantRes[idx].body, string(resp))
				}
			}
			time.Sleep(10 * time.Millisecond)
			log.Println("shutdown: ", g.ShutDown())
		})
	}
}