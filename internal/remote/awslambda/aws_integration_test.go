//go:build integration

package awslambda

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testRegion = "us-east-1"

func startMiniStack(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nahuelnucera/ministack:latest",
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForListeningPort("4566/tcp").WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "4566/tcp")
	require.NoError(t, err)
	return fmt.Sprintf("http://%s:%s", host, port.Port())
}

func newTestAwsConfig(t *testing.T, endpoint string) aws.Config {
	t.Helper()
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(testRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	require.NoError(t, err)
	return cfg
}

func newS3Client(cfg aws.Config, endpoint string) *s3.Client {
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

func newLambdaClient(cfg aws.Config, endpoint string) *lambda.Client {
	return lambda.NewFromConfig(cfg, func(o *lambda.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
}

func newIAMClient(cfg aws.Config, endpoint string) *iam.Client {
	return iam.NewFromConfig(cfg, func(o *iam.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
}

func TestS3BucketOperations(t *testing.T) {
	endpoint := startMiniStack(t)
	cfg := newTestAwsConfig(t, endpoint)
	svc := newS3Client(cfg, endpoint)

	bucketName := "test-imposter-bucket"

	t.Run("create bucket", func(t *testing.T) {
		err := ensureBucket(svc, bucketName, testRegion)
		require.NoError(t, err)
	})

	t.Run("upload object", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test-upload.txt"
		require.NoError(t, os.WriteFile(tmpFile, []byte("hello ministack"), 0644))

		err := upload(svc, bucketName, tmpFile, "test-key.txt")
		require.NoError(t, err)
	})

	t.Run("ensure bucket idempotent", func(t *testing.T) {
		err := ensureBucket(svc, bucketName, testRegion)
		require.NoError(t, err)
	})
}

func TestIAMRoleCreation(t *testing.T) {
	endpoint := startMiniStack(t)
	cfg := newTestAwsConfig(t, endpoint)
	iamSvc := newIAMClient(cfg, endpoint)

	roleName := "TestImposterRole"

	t.Run("create role", func(t *testing.T) {
		roleArn, err := createRoleWithClient(iamSvc, roleName)
		require.NoError(t, err)
		require.NotEmpty(t, roleArn)
		t.Logf("created role ARN: %s", roleArn)
	})

	t.Run("get existing role", func(t *testing.T) {
		result, err := iamSvc.GetRole(context.TODO(), &iam.GetRoleInput{
			RoleName: &roleName,
		})
		require.NoError(t, err)
		require.NotNil(t, result.Role.Arn)
		t.Logf("got role ARN: %s", *result.Role.Arn)
	})
}

func TestLambdaFunctionLifecycle(t *testing.T) {
	endpoint := startMiniStack(t)
	cfg := newTestAwsConfig(t, endpoint)
	svc := newLambdaClient(cfg, endpoint)
	iamSvc := newIAMClient(cfg, endpoint)

	roleName := "TestLambdaRole"
	roleArn, err := createRoleWithClient(iamSvc, roleName)
	require.NoError(t, err)

	funcName := "test-imposter-func"
	dummyZip := createDummyZip(t)

	var funcArn string

	t.Run("create function", func(t *testing.T) {
		funcArn, err = createFunction(
			svc,
			testRegion,
			funcName,
			roleArn,
			256,
			LambdaArchitectureX86_64,
			codeLocation{zipContents: &dummyZip},
			false,
		)
		require.NoError(t, err)
		require.NotEmpty(t, funcArn)
		t.Logf("created function ARN: %s", funcArn)
	})

	t.Run("check function exists", func(t *testing.T) {
		result, err := checkFunctionExists(svc, funcName)
		require.NoError(t, err)
		require.NotNil(t, result.Configuration)
		require.Equal(t, funcName, *result.Configuration.FunctionName)
	})

	t.Run("update function code", func(t *testing.T) {
		err := updateFunctionCode(svc, funcArn, codeLocation{zipContents: &dummyZip})
		require.NoError(t, err)
	})

	t.Run("delete function", func(t *testing.T) {
		_, err := svc.DeleteFunction(context.TODO(), &lambda.DeleteFunctionInput{
			FunctionName: aws.String(funcArn),
		})
		require.NoError(t, err)
	})
}

func TestLambdaFunctionUrlConfig(t *testing.T) {
	endpoint := startMiniStack(t)
	cfg := newTestAwsConfig(t, endpoint)
	svc := newLambdaClient(cfg, endpoint)
	iamSvc := newIAMClient(cfg, endpoint)

	roleName := "TestUrlRole"
	roleArn, err := createRoleWithClient(iamSvc, roleName)
	require.NoError(t, err)

	funcName := "test-url-func"
	dummyZip := createDummyZip(t)
	funcArn, err := createFunction(
		svc,
		testRegion,
		funcName,
		roleArn,
		256,
		LambdaArchitectureX86_64,
		codeLocation{zipContents: &dummyZip},
		false,
	)
	require.NoError(t, err)

	t.Run("create function URL config", func(t *testing.T) {
		result, err := svc.CreateFunctionUrlConfig(context.TODO(), &lambda.CreateFunctionUrlConfigInput{
			AuthType:     lambdatypes.FunctionUrlAuthTypeNone,
			FunctionName: aws.String(funcArn),
		})
		require.NoError(t, err)
		require.NotNil(t, result.FunctionUrl)
		t.Logf("function URL: %s", *result.FunctionUrl)
	})

	t.Run("configure anonymous access", func(t *testing.T) {
		err := configureUrlAccess(svc, funcArn, true)
		require.NoError(t, err)
	})

	t.Run("remove anonymous access", func(t *testing.T) {
		err := configureUrlAccess(svc, funcArn, false)
		require.NoError(t, err)
	})
}

// createRoleWithClient creates an IAM role using the provided IAM client.
func createRoleWithClient(svc *iam.Client, roleName string) (string, error) {
	assumeRolePolicy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`
	result, err := svc.CreateRole(context.TODO(), &iam.CreateRoleInput{
		RoleName:                 &roleName,
		AssumeRolePolicyDocument: &assumeRolePolicy,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create role %s: %v", roleName, err)
	}
	return *result.Role.Arn, nil
}

// createDummyZip creates a minimal valid zip file for Lambda function creation.
func createDummyZip(t *testing.T) []byte {
	t.Helper()
	// Minimal valid zip file (empty archive)
	return []byte{
		0x50, 0x4b, 0x05, 0x06, // end of central directory signature
		0x00, 0x00, // number of this disk
		0x00, 0x00, // disk where central directory starts
		0x00, 0x00, // number of central directory records on this disk
		0x00, 0x00, // total number of central directory records
		0x00, 0x00, 0x00, 0x00, // size of central directory
		0x00, 0x00, 0x00, 0x00, // offset of start of central directory
		0x00, 0x00, // comment length
	}
}
