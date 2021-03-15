package main

import (
	"context"
	"fmt"
	"github.com/muerwre/vault-golang/internal/api"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "vault",
	Short: "vault backend",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func serveAPI(ctx context.Context, api *api.API) {
	router := api.Init()

	hasCerts := len(api.Config.TlsFiles) == 2

	s := &http.Server{
		Addr:        fmt.Sprintf(":%d", api.Config.Port),
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
		logrus.Infof("Https listening at port %d", api.Config.Port)

		if err := s.ListenAndServeTLS(api.Config.TlsFiles[0], api.Config.TlsFiles[1]); err != http.ErrServerClosed {
			logrus.Error(err)
		}
	} else {
		logrus.Infof("Http listening at port %d", api.Config.Port)

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
			go a.Mailer.Listen(ctx)
			go a.Notifier.Listen(ctx)
			go a.Vk.Watch(ctx)
			<-ctx.Done()
			a.Close()
		}()

		wg.Wait()

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Warnf(err.Error())
		os.Exit(1)
	}
}

var configFile string

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.Warnf("Unable to read config: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is config.yaml)")
	rootCmd.AddCommand(serveCmd)
	Execute()
}
