package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"my_tristack_app/internal/config"
	"my_tristack_app/internal/db"
	"my_tristack_app/internal/handler"
	"my_tristack_app/internal/repository"
	"my_tristack_app/internal/service"
	"my_tristack_app/internal/web"
)

func main() {
	cfg := config.Load()

	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("database not ready: %v", err)
	}
	repo := repository.NewItemRepository(conn)
	items := service.NewItemService(repo)
	h := handler.NewHandler(items)

	webSrv, err := web.New(cfg.AppName, items)
	if err != nil {
		log.Fatalf("web templates: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/api/v1/items", h.ListItems)
	r.POST("/api/v1/items", h.CreateItem)
	webHandler := gin.WrapH(webSrv.Mux())
	r.GET("/", webHandler)
	r.GET("/web/items", webHandler)
	r.GET("/static/*filepath", webHandler)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("%s listening on %s", cfg.AppName, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Printf("server stopped")
}
