package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"istio-rest-demo/internal"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	var port int
	var basePath string
	flag.IntVar(&port, "p", 18081, "http server port")
	flag.StringVar(&basePath, "b", "/server", "http server base path")
	flag.Parse()
	r := gin.New()
	ginLoggerConfig := gin.LoggerConfig{SkipPaths: []string{"/actuator/health"}}
	r.Use(gin.LoggerWithConfig(ginLoggerConfig), gin.Recovery())
	r.UseH2C = true
	g1 := r.Group(basePath)
	{
		g1.GET("hello", func(c *gin.Context) {
			log.Println("-------------- server request headers: --------------")
			for k, v := range c.Request.Header {
				log.Printf("%s : %s\n", k, v)
			}
			var params internal.DemoParams
			_ = c.ShouldBind(&params)
			if len(params.Pause) > 0 {
				d1, err := time.ParseDuration(params.Pause)
				if err == nil {
					time.Sleep(d1)
				}
			}
			if strings.EqualFold("tom", params.Name) {
				c.Header(internal.ErrorReason, url.QueryEscape("名称不能为 tom !"))
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}
			id, _ := uuid.NewV7()
			result := internal.JsonResult[[]string]{
				Code:    http.StatusOK,
				Message: fmt.Sprintf("port : %d hello %s -> %s", port, params.Name, id),
				Data:    getAddrs(),
			}
			c.JSON(http.StatusOK, result)
		})
	}
	g2 := r.Group("actuator")
	{
		g2.GET("health", func(c *gin.Context) {
			id, _ := uuid.NewV7()
			result := internal.JsonResult[string]{
				Code:    http.StatusOK,
				Message: fmt.Sprintf("health port : %d => %s", port, id),
				Data:    time.Now().Format(time.RFC3339),
			}
			c.JSON(http.StatusOK, result)
		})
	}
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Panic(err)
	}
}

func getAddrs() []string {
	addrs, err := net.InterfaceAddrs()
	var r []string
	if err != nil {
		return r
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				r = append(r, ipnet.IP.String())
			}
		}
	}
	return r
}
