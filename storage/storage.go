package storage

import (
	"fmt"
	"math/rand"

	"github.com/lnsp/cloudsort/pb"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func GetS3Client(creds *pb.S3Credentials) (*minio.Client, error) {
	// Just pick a random endpoint
	ep := creds.Endpoints[rand.Intn(len(creds.Endpoints))]
	client, err := minio.New(ep, &minio.Options{
		Region: creds.Region,
		Secure: !creds.DisableSsl,
		Creds:  credentials.NewStaticV4(creds.AccessKeyId, creds.SecretAccessKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("create S3 client: %w", err)
	}
	return client, nil
}
