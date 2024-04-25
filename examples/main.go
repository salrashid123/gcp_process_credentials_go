package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	sal "github.com/salrashid123/gcp_process_credentials_go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"golang.org/x/oauth2"
)

var (
	projectID  = "your_project"
	bucketName = "your_bucket"
)

func main() {

	ctx := context.Background()

	extTokenSource, err := sal.ExternalTokenSource(
		&sal.ExternalTokenConfig{
			Command: "/usr/bin/cat",
			Env:     []string{"foo=bar"},
			Args:    []string{"/tmp/token.txt"},
		},
	)

	tok, err := extTokenSource.Token()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Token Token: %s\n", tok.AccessToken)

	sts := oauth2.ReuseTokenSource(tok, extTokenSource)

	storageClient, err := storage.NewClient(ctx, option.WithTokenSource(sts))
	if err != nil {
		log.Fatalf("Could not create storage Client: %v", err)
	}

	it := storageClient.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			panic(err)
		}
		log.Printf("Bucket: %v\n", battrs.Name)
	}
}
