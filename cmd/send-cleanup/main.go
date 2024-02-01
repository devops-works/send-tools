package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"cloud.google.com/go/storage"
	"github.com/carlmjohnson/versioninfo"
	"github.com/redis/go-redis/v9"

	"google.golang.org/api/iterator"
)

const BUCKET = "dw-hub-prod-35cfbedc-dwk-send-gcs"

func main() {
	var (
		redis       string
		bucket      string
		forReal     bool
		notInRedis  []string
		notInBucket []string
	)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("send-cleanup", "version", versioninfo.Version, "revision", versioninfo.Revision, "dirty", versioninfo.DirtyBuild, "lastcommit", versioninfo.LastCommit)
	if isInSlice(os.Args, "-v") {
		os.Exit(0)
	}

	flag.StringVar(&redis, "r", "127.0.0.1:6379", "Redis server address")
	flag.StringVar(&bucket, "b", "", "GCS bucket name (mandatory)")
	flag.BoolVar(&forReal, "y", false, "Execute deletes")
	flag.Parse()

	if bucket == "" {
		logger.Error("bucket must be specified")
		flag.Usage()
		os.Exit(1)
	}
	ctx := context.TODO()

	// list objects in gcs bucket
	logger.Info("analyzing bucket", "bucket", bucket)
	files, err := listObjects(ctx, bucket)
	if err != nil {
		logger.Error("unable to check bucket content", "error", err)
		os.Exit(1)
	}

	// list redis keys with hgetall
	logger.Info("analyzing redis", "server", redis)
	redisKeys, err := listRedisKeys(ctx, logger, redis)
	if err != nil {
		logger.Error("unable to check redis content", "error", err)
		os.Exit(1)
	}

	for _, key := range redisKeys {
		// check if key is present on GCS
		if !isInSlice(files, key) {
			logger.Info("key not in bucket", "key", key)
			notInBucket = append(notInBucket, key)
		}
	}

	for _, file := range files {
		// check if file is listed in redisKeys
		if !isInSlice(redisKeys, file) {
			logger.Info("file not in redis", "file", file)
			notInRedis = append(notInRedis, file)
		}
	}

	logger.Info("files in bucket", "count", len(files))
	logger.Info("redis keys", "count", len(redisKeys))
	logger.Info("keys not in bucket", "count", len(notInBucket))
	logger.Info("files not in redis", "count", len(notInRedis))

	if !forReal {
		logger.Info("dry run, exiting")
		os.Exit(0)
	}

	logger.Warn("removing keys from redis", "count", len(notInBucket))
	removeRedisKeys(ctx, logger, redis, notInBucket)

	logger.Info("removing files from GCS", "count", len(notInRedis))
	removeGCSObjects(ctx, logger, bucket, notInRedis)
}

func removeRedisKeys(ctx context.Context, logger *slog.Logger, srv string, keys []string) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     srv,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	for _, key := range keys {
		// remove key from redis
		if err := rdb.Del(ctx, key).Err(); err != nil {
			logger.Error("unable to remove key from redis", "key", key, "error", err)
		}
	}
}

func removeGCSObjects(ctx context.Context, logger *slog.Logger, bucket string, keys []string) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		logger.Error("unable to create GCS client", "error", err)
	}
	defer client.Close()

	for _, key := range keys {
		// remove file from GCS
		if err := client.Bucket(bucket).Object(key).Delete(ctx); err != nil {
			logger.Error("unable to remove key from GCS", "bucket", bucket, "key", key, "error", err)
		}
	}
}

func isInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func listRedisKeys(ctx context.Context, logger *slog.Logger, srv string) ([]string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     srv,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// iterate over all keys
	iter := rdb.ScanType(ctx, 0, "", 0, "hash").Iterator()
	var keys []string
	for iter.Next(ctx) {
		// hgetall key
		val, err := rdb.HGetAll(ctx, iter.Val()).Result()
		if err != nil {
			logger.Error("unable to get key", "key", iter.Val(), "error", err)
			continue
		}
		keys = append(keys, val["prefix"]+"-"+iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func listObjects(ctx context.Context, bucket string) ([]string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var files []string
	it := client.Bucket(bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, attrs.Name)
	}
	return files, nil
}
