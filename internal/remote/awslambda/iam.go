package awslambda

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const defaultIamRoleName = "ImposterLambdaExecutionRole"

func ensureIamRole(session *session.Session, roleName string) (string, error) {
	svc := iam.New(session)
	getRoleResult, err := svc.GetRole(&iam.GetRoleInput{
		RoleName: &roleName,
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
				roleArn, err := createRole(svc, roleName)
				if err != nil {
					return "", err
				}
				return roleArn, nil

			} else {
				logger.Fatalf("failed to get IAM role: %s: %v", roleName, err)
			}
		} else {
			logger.Fatal(err)
		}
	}
	logger.Debugf("using role: %s", *getRoleResult.Role.Arn)
	return *getRoleResult.Role.Arn, nil
}

func createRole(svc *iam.IAM, roleName string) (string, error) {
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
	createRoleOutput, err := svc.CreateRole(&iam.CreateRoleInput{
		Description:              &description,
		RoleName:                 &roleName,
		AssumeRolePolicyDocument: &assumeRolePolicy,
	})
	if err != nil {
		logger.Fatalf("failed to create role: %s: %v", roleName, err)
	}
	roleArn := *createRoleOutput.Role.Arn

	arn := "arn:aws:iam::aws:policy/AWSLambdaExecute"
	getPolicyResult, err := svc.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: &arn,
	})
	if err != nil {
		return "", err
	}
	_, err = svc.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: getPolicyResult.Policy.Arn,
		RoleName:  &roleName,
	})
	if err != nil {
		return "", err
	}
	logger.Debugf("created role: %s with arn: %s", roleName, roleArn)
	return roleArn, nil
}
