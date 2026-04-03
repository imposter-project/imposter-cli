package awslambda

import (
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

const defaultIamRoleName = "ImposterLambdaExecutionRole"

func ensureIamRole(cfg aws.Config, roleName string) (string, error) {
	svc := iam.NewFromConfig(cfg)
	getRoleResult, err := svc.GetRole(ctx, &iam.GetRoleInput{
		RoleName: &roleName,
	})
	if err != nil {
		var notFoundErr *iamtypes.NoSuchEntityException
		if errors.As(err, &notFoundErr) {
			roleArn, err := createRole(svc, roleName)
			if err != nil {
				return "", err
			}
			return roleArn, nil
		} else {
			logger.Fatalf("failed to get IAM role: %s: %v", roleName, err)
		}
	}
	logger.Debugf("using role: %s", *getRoleResult.Role.Arn)
	return *getRoleResult.Role.Arn, nil
}

func createRole(svc *iam.Client, roleName string) (string, error) {
	description := "Default IAM role for Imposter Lambda"
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
	createRoleOutput, err := svc.CreateRole(ctx, &iam.CreateRoleInput{
		Description:              &description,
		RoleName:                 &roleName,
		AssumeRolePolicyDocument: &assumeRolePolicy,
	})
	if err != nil {
		logger.Fatalf("failed to create role: %s: %v", roleName, err)
	}
	roleArn := *createRoleOutput.Role.Arn

	arn := "arn:aws:iam::aws:policy/AWSLambdaExecute"
	getPolicyResult, err := svc.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: &arn,
	})
	if err != nil {
		return "", err
	}
	_, err = svc.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		PolicyArn: getPolicyResult.Policy.Arn,
		RoleName:  &roleName,
	})
	if err != nil {
		return "", err
	}
	logger.Debugf("created role: %s with arn: %s", roleName, roleArn)
	return roleArn, nil
}
