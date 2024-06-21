package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"istio-rest-demo/internal"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	var port int
	flag.IntVar(&port, "p", 18113, "http server port")
	var basePath string
	flag.StringVar(&basePath, "b", "/client", "http server base path")
	var targetAddr string
	flag.StringVar(&targetAddr, "t", "http://192.168.0.188:10000/server", "service provider addr")
	flag.Parse()
	restClient := resty.New().SetBaseURL(targetAddr)
	r := gin.New()
	ginLoggerConfig := gin.LoggerConfig{SkipPaths: []string{"/actuator/health"}}
	r.Use(gin.LoggerWithConfig(ginLoggerConfig), gin.Recovery())
	r.UseH2C = true
	g1 := r.Group(basePath)
	{
		g1.GET("hello", func(c *gin.Context) {
			params := struct {
				internal.DemoParams
				EnableTimeout string `json:"enableTimeout" form:"enableTimeout"`
				EnableRetry   string `json:"enableRetry" form:"enableRetry"`
				EnvoyTimeout  string `json:"envoyTimeout" form:"envoyTimeout"`
			}{}
			_ = c.ShouldBind(&params)
			var result internal.JsonResult[[]string]
			err := internal.Breaker(func() (*resty.Response, error) {
				r1 := restClient.R().
					SetResult(&result).
					SetQueryParam("name", params.Name).
					SetQueryParam("pause", params.Pause)
				if len(params.EnableTimeout) > 0 {
					r1.SetHeader("enable-timeout", params.EnableTimeout)
				}
				if len(params.EnableRetry) > 0 {
					r1.SetHeader("enable-retry", params.EnableRetry)
				}
				if len(params.EnvoyTimeout) > 0 {
					d1, err := time.ParseDuration(params.EnvoyTimeout)
					if err == nil {
						ms := strconv.FormatInt(d1.Milliseconds(), 10)
						r1.SetHeader("x-envoy-upstream-rq-timeout-ms", ms)
						r1.SetHeader("x-envoy-upstream-rq-per-try-timeout-ms", ms)
						r1.SetHeader("x-envoy-hedge-on-per-try-timeout", "true")
					}
				}
				log.Println("-------------- client request headers: --------------")
				for k, v := range r1.Header {
					log.Printf("%s : %s\n", k, v)
				}
				return r1.Get("/hello")
			})
			if err == nil {
				c.JSON(http.StatusOK, result)
			} else {
				c.JSON(http.StatusInternalServerError,
					internal.JsonResult[any]{
						Code:  http.StatusInternalServerError,
						Error: err.Error(),
					},
				)
			}
		})
		g1.GET("holiday", func(c *gin.Context) {
			params := struct {
				Year int `json:"year" form:"year"`
			}{}
			_ = c.ShouldBind(&params)
			result := struct {
				Code    int                    `json:"code" form:"code"`
				Holiday map[string]interface{} `json:"holiday" form:"holiday"`
			}{}
			err := internal.Breaker(func() (*resty.Response, error) {
				return restClient.R().
					SetResult(&result).
					SetPathParam("year", strconv.Itoa(params.Year)).
					Get("https://timor.tech/api/holiday/year/{year}?type=Y&week=N")
			})
			if err == nil {
				c.JSON(http.StatusOK, result)
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": err.Error(),
				})
			}
		})
	}
	g2 := r.Group("actuator")
	{
		g2.GET("health", func(c *gin.Context) {
			id, _ := uuid.NewV7()
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": fmt.Sprintf("health port : %d => %s", port, id),
				"data":    time.Now().Format(time.RFC3339),
			})
		})
	}
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Panic(err)
	}
}
