package awslambda

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/imposter-project/imposter-cli/internal/stringutil"
	"os"
	"strings"
)

const defaultS3ObjectKey = "imposter-bundle.zip"

func (m LambdaRemote) uploadBundleToBucket(zipContents *[]byte) (bucketName string, objectKey string, err error) {
	localBundlePath, err := m.writeBundleToTempFile(zipContents)
	if err != nil {
		return "", "", err
	}
	bucketName, err = m.getBucketName()
	if err != nil {
		return "", "", err
	}
	objectKey = stringutil.GetFirstNonEmpty(m.Config[configKeyS3ObjectKey], defaultS3ObjectKey)
	if err = m.uploadToBucket(localBundlePath, bucketName, objectKey); err != nil {
		return "", "", fmt.Errorf("failed to upload file %v to bucket %v: %v", localBundlePath, bucketName, err)
	}
	return bucketName, objectKey, nil
}

func (m LambdaRemote) writeBundleToTempFile(zipContents *[]byte) (localBundlePath string, err error) {
	temp, err := os.CreateTemp(os.TempDir(), "imposter-bundle-*.zip")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer temp.Close()

	localBundlePath = temp.Name()
	if err = os.WriteFile(localBundlePath, *zipContents, 0644); err != nil {
		return "", fmt.Errorf("failed to write bundle to temp file %v: %v", temp, err)
	}
	logger.Tracef("wrote bundle to temp file: %v", localBundlePath)
	return localBundlePath, nil
}

func (m LambdaRemote) getBucketName() (bucketName string, err error) {
	bucketName = m.Config[configKeyS3BucketName]
	if bucketName == "" {
		bucketName = "imposter-mock-" + strings.ReplaceAll(uuid.New().String(), "-", "")
		m.Config[configKeyS3BucketName] = bucketName
		if err = m.SaveConfig(); err != nil {
			return "", fmt.Errorf("failed to save bucket name %v in config: %v", bucketName, err)
		}
	}
	return bucketName, nil
}

func (m LambdaRemote) uploadToBucket(localPath string, bucketName string, objectKey string) error {
	region, svc, err := m.initS3Client()
	if err != nil {
		return fmt.Errorf("failed to initialise S3 client: %v", err)
	}
	if err = ensureBucket(svc, bucketName, region); err != nil {
		return fmt.Errorf("failed to ensure bucket %v exists: %v", bucketName, err)
	}
	if err = upload(svc, bucketName, localPath, objectKey); err != nil {
		return fmt.Errorf("failed to upload file %v to bucket %v: %v", localPath, bucketName, err)
	}
	return nil
}

func ensureBucket(svc *s3.Client, bucketName string, region string) error {
	logger.Tracef("checking for bucket %v in region %v", bucketName, region)

	if _, err := svc.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucketName)}); err != nil {
		if err = createBucket(svc, bucketName, region); err != nil {
			return err
		}
	}
	logger.Debugf("bucket %v exists", bucketName)
	return nil
}

func createBucket(svc *s3.Client, bucketName string, region string) error {
	logger.Tracef("creating bucket %v in region %v", bucketName, region)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	// us-east-1 must not specify LocationConstraint
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}

	_, err := svc.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create bucket %v in region %v: %v", bucketName, region, err)
	}
	logger.Debugf("created bucket %v in region %v", bucketName, region)
	return nil
}

func upload(svc *s3.Client, bucketName string, localPath string, objectKey string) error {
	logger.Tracef("uploading file %v to bucket %v", localPath, bucketName)

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v: %v", localPath, err)
	}
	defer file.Close()

	_, err = svc.PutObject(ctx, &s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %v to bucket %v: %v", localPath, bucketName, err)
	}
	logger.Debugf("uploaded file %v to bucket %v", localPath, bucketName)
	return nil
}

func (m LambdaRemote) initS3Client() (region string, svc *s3.Client, err error) {
	if m.Config[configKeyRegion] == "" {
		return "", nil, fmt.Errorf("region cannot be null")
	}
	region, cfg, err := m.loadAwsConfig()
	if err != nil {
		return "", nil, err
	}
	svc = s3.NewFromConfig(cfg)
	return region, svc, nil
}
