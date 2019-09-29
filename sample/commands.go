package main

import (
	"github.com/spf13/cobra"

	"github.com/undeconstructed/sample/auth"
	"github.com/undeconstructed/sample/config"
	"github.com/undeconstructed/sample/fetcher"
	"github.com/undeconstructed/sample/frontend"
	"github.com/undeconstructed/sample/store"
	"github.com/undeconstructed/sample/user"
)

func makeTestMode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testmode",
		Short: "Test",
		Long:  `test.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runService(makeTestService())
		},
	}

	return cmd
}

func makeRunAuth() *cobra.Command {
	var httpBind, userURL string

	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Auth service",
		Long:  `test.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runService(auth.New(httpBind, userURL))
		},
	}
	cmd.Flags().StringVarP(&httpBind, "http-bind", "", ":8080", "where to bind HTTP")
	cmd.Flags().StringVarP(&userURL, "user", "", "", "user URL")
	cmd.MarkFlagRequired("user")

	return cmd
}

func makeRunConfig() *cobra.Command {
	var grpcBind, httpBind, path, defaultStoreURL string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Config service",
		Long:  `test.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runService(config.New(grpcBind, httpBind, path, defaultStoreURL))
		},
	}
	cmd.Flags().StringVarP(&grpcBind, "grpc-bind", "", ":8000", "where to bind gRPC")
	cmd.Flags().StringVarP(&httpBind, "http-bind", "", ":8080", "where to bind HTTP")
	cmd.Flags().StringVarP(&path, "data-path", "", "config.json", "data storage path")
	cmd.Flags().StringVarP(&defaultStoreURL, "default-store", "", "", "default store")
	cmd.MarkFlagRequired("default-store")

	return cmd
}

func makeRunFetcher() *cobra.Command {
	var configURL string

	cmd := &cobra.Command{
		Use:   "fetcher",
		Short: "Fetcher service",
		Long:  `test.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runService(fetcher.New(configURL))
		},
	}
	cmd.Flags().StringVarP(&configURL, "config", "", "", "config URL")
	cmd.MarkFlagRequired("config")

	return cmd
}

func makeRunFrontend() *cobra.Command {
	var httpBind, configURL, userURL string

	cmd := &cobra.Command{
		Use:   "frontend",
		Short: "Frontend service",
		Long:  `test.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runService(frontend.New(httpBind, configURL, userURL))
		},
	}
	cmd.Flags().StringVarP(&httpBind, "http-bind", "", ":8080", "where to bind HTTP")
	cmd.Flags().StringVarP(&configURL, "config", "", "", "config URL")
	cmd.MarkFlagRequired("config")
	cmd.Flags().StringVarP(&configURL, "user", "", "", "user URL")
	cmd.MarkFlagRequired("user")

	return cmd
}

func makeRunStore() *cobra.Command {
	var grpcBind, path string

	cmd := &cobra.Command{
		Use:   "store",
		Short: "Store service",
		Long:  `test.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runService(store.New(grpcBind, path))
		},
	}
	cmd.Flags().StringVarP(&grpcBind, "grpc-bind", "", ":8000", "where to bind gRPC")
	cmd.Flags().StringVarP(&path, "data-path", "", "store.db", "data storage path")

	return cmd
}

func makeRunUser() *cobra.Command {
	var grpcBind string

	cmd := &cobra.Command{
		Use:   "user",
		Short: "User service",
		Long:  `test.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runService(user.New(grpcBind))
		},
	}
	cmd.Flags().StringVarP(&grpcBind, "grpc-bind", "", ":8000", "where to bind gRPC")

	return cmd
}
