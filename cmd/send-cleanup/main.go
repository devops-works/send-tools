package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/devops-works/send-tools/internal/backends"
	"github.com/devops-works/send-tools/internal/backends/gcs"
	"github.com/devops-works/send-tools/internal/backends/redis"
)

var (
	version   string
	buildDate string
)

func main() {
	var (
		redSrv  string
		path    string
		forReal bool

		storage backends.Storage
	)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("send-cleanup", "built", buildDate, "version", version)
	if isInSlice(os.Args, "-v") {
		os.Exit(0)
	}

	flag.StringVar(&redSrv, "r", "127.0.0.1:6379", "Redis server address")
	flag.StringVar(&path, "p", "", "Path to storage (mandatory)")
	flag.BoolVar(&forReal, "y", false, "Execute deletes")
	flag.Parse()

	if path == "" {
		logger.Error("path must be specified")
		flag.Usage()
		os.Exit(1)
	}

	switch {
	case path[:5] == "gs://":
		// GCS
		logger.Info("using GCS")
		bucket := path[5:]
		storage = gcs.NewGCS(bucket)

	case path[:5] == "s3://":
		// S3
		logger.Info("using S3")
		logger.Error("S3 not implemented")
		os.Exit(1)
		// bucket := path[5:]
		// storage = s3.NewS3(bucket)

	case path[:1] == "/":
		// Filesystem
		logger.Info("using filesystem")
		logger.Error("filesystem not implemented")
		os.Exit(1)
		// storage = filesystem.NewFilesystem(path)

	default:
		logger.Error("unsupported storage", "path", path)
		os.Exit(1)
	}

	rd := redis.NewRedis(redSrv)

	err := run(logger, rd, storage, forReal)
	if err != nil {
		logger.Error("unable to run cleanup", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, rd *redis.Redis, storage backends.Storage, forReal bool) error {
	var (
		notInRedis []string
		notInpath  []string
	)

	ctx := context.TODO()

	// list objects in gcs path
	logger.Info("analyzing path", "path", storage.Path())
	files, err := storage.ListObjects(ctx, storage.Path())
	if err != nil {
		logger.Error("unable to check path content", "error", err)
		os.Exit(1)
	}

	// list redis keys with hgetall
	logger.Info("analyzing redis", "server", rd.Server)
	redisKeys, err := rd.ListObjects(ctx)
	if err != nil {
		logger.Error("unable to check redis content", "error", err)
		os.Exit(1)
	}

	for _, key := range redisKeys {
		// check if redis key is present on storage
		if !isInSlice(files, key) {
			logger.Info("key not in storage", "key", key)
			notInpath = append(notInpath, key)
		}
	}

	for _, file := range files {
		// check if file is listed in redisKeys
		if !isInSlice(redisKeys, file) {
			logger.Info("file not in redis", "file", file)
			notInRedis = append(notInRedis, file)
		}
	}

	logger.Info("files in path", "count", len(files))
	logger.Info("redis keys", "count", len(redisKeys))
	logger.Info("keys not in path", "count", len(notInpath))
	logger.Info("files not in redis", "count", len(notInRedis))

	if !forReal {
		logger.Info("dry run, exiting")
		os.Exit(0)
	}

	logger.Warn("removing keys from redis", "count", len(notInpath))
	for _, key := range notInpath {
		if err := rd.RemoveObject(ctx, key); err != nil {
			logger.Error("unable to remove key from redis", "key", key, "error", err)
		}
	}

	logger.Info("removing files from GCS", "count", len(notInRedis))
	for _, file := range notInRedis {
		if err := storage.RemoveObject(ctx, file); err != nil {
			logger.Error("unable to remove file from storage", "file", file, "error", err)
		}
	}

	return nil
}

func isInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
