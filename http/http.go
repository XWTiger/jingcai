package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"jingcai/config"
	"jingcai/router"
	"net/http"
	"time"
)

func Init(c *config.Config) func() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	//bind router
	router.BindRouters(r, c)
	addr := fmt.Sprintf("%s:%d", c.HttpConf.Host, c.HttpConf.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(c.HttpConf.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.HttpConf.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(c.HttpConf.IdleTimeout) * time.Second,
	}
	go func() {
		fmt.Println("http服务启动:", addr)
		var err error
		if c.HttpConf.CertFile != "" && c.HttpConf.KeyFile != "" {
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(c.HttpConf.CertFile, c.HttpConf.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(c.HttpConf.ShutdownTimeout))
		defer cancel()
		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Println("http服务关闭错误:", err)
		}

		select {
		case <-ctx.Done():
			fmt.Println("http服务退出")
		default:
			fmt.Println("http服务关闭")
		}
	}
}
