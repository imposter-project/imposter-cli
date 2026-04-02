package awslambda

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func configureUrlAccess(svc *lambda.Client, funcArn string, anonAccess bool) error {
	const statementId = "PermitAnonymousAccessToFunctionUrl"
	if anonAccess {
		if err := createAnonUrlAccessPolicy(svc, funcArn, statementId); err != nil {
			return err
		}
	} else {
		_, err := svc.RemovePermission(context.TODO(), &lambda.RemovePermissionInput{
			FunctionName: aws.String(funcArn),
			StatementId:  aws.String(statementId),
		})
		if err != nil {
			var notFoundErr *lambdatypes.ResourceNotFoundException
			if errors.As(err, &notFoundErr) {
				logger.Debugf("anonymous URL access permission did not exist")
				return nil
			}
			return fmt.Errorf("failed to delete anonymous URL access permission: %v", err)
		}
		logger.Debugf("deleted anonymous URL access permission")
	}
	return nil
}

func createAnonUrlAccessPolicy(svc *lambda.Client, funcArn string, statementId string) error {
	_, err := svc.AddPermission(context.TODO(), &lambda.AddPermissionInput{
		StatementId:         aws.String(statementId),
		Action:              aws.String("lambda:InvokeFunctionUrl"),
		FunctionName:        aws.String(funcArn),
		FunctionUrlAuthType: lambdatypes.FunctionUrlAuthTypeNone,
		Principal:           aws.String("*"),
	})
	if err != nil {
		var conflictErr *lambdatypes.ResourceConflictException
		if errors.As(err, &conflictErr) {
			logger.Debugf("anonymous URL access permission already exists")
			return nil
		}
		return fmt.Errorf("failed to add anonymous URL access permission: %v", err)
	}
	logger.Debugf("added anonymous URL access permission")
	return nil
}

func (m LambdaRemote) ensureUrlConfigured(svc *lambda.Client, funcArn string) (string, error) {
	logger.Debugf("configuring URL for function: %s", funcArn)

	var functionUrl string
	getUrlResult, err := m.checkFunctionUrlConfig(svc, funcArn)
	if err != nil {
		var notFoundErr *lambdatypes.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			urlConfigOutput, err := svc.CreateFunctionUrlConfig(context.TODO(), &lambda.CreateFunctionUrlConfigInput{
				AuthType:     lambdatypes.FunctionUrlAuthTypeNone,
				FunctionName: aws.String(funcArn),
			})
			if err != nil {
				return "", fmt.Errorf("failed to create URL for function: %s: %v", funcArn, err)
			}
			functionUrl = *urlConfigOutput.FunctionUrl
			logger.Debugf("configured function URL: %s", functionUrl)
		} else {
			return "", fmt.Errorf("failed to check if URL config exists for function: %s: %v", funcArn, err)
		}
	} else {
		functionUrl = *getUrlResult.FunctionUrl
		logger.Debugf("function URL already configured: %s", functionUrl)
	}
	return functionUrl, nil
}
