package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/api"
	"github.com/muerwre/vault-golang/app"
)

func serveAPI(ctx context.Context, api *api.API) {
	router := gin.Default()
	// router.LoadHTMLGlob("views/*")
	api.Init(router.Group("/"))

	hasCerts := len(api.App.Config.TlsFiles) == 2

	s := &http.Server{
		Addr:        fmt.Sprintf("%s:%d", api.App.Config.Host, api.App.Config.Port),
		Handler:     router,
		ReadTimeout: 2 * time.Minute,
	}

	done := make(chan struct{})

	go func() {
		<-ctx.Done()

		if err := s.Shutdown(context.Background()); err != nil {
			logrus.Error(err)
		}

		close(done)
	}()

	if hasCerts {
		logrus.Infof(fmt.Sprintf("Listening https://%s:%d", api.App.Config.Host, api.App.Config.Port))

		if err := s.ListenAndServeTLS(api.App.Config.TlsFiles[0], api.App.Config.TlsFiles[1]); err != http.ErrServerClosed {
			logrus.Error(err)
		}
	} else {
		logrus.Infof(fmt.Sprintf("Listening http://%s:%d", api.App.Config.Host, api.App.Config.Port))

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Error(err)
		}
	}

	<-done
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serves the api",
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := app.New()

		if err != nil {
			return err
		}

		defer a.Close()

		api, err := api.New(a)

		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, os.Interrupt)
			<-ch
			logrus.Info("signal caught. shutting down...")
			cancel()
		}()

		var wg sync.WaitGroup

		wg.Add(2)

		go func() {
			defer wg.Done()
			defer cancel()
			serveAPI(ctx, api)
		}()

		go func() {
			defer wg.Done()
			go a.Mailer.Listen()
			<-ctx.Done()
			a.Mailer.Done()
		}()

		wg.Wait()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
