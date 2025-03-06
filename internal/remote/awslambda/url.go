package awslambda

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func configureUrlAccess(svc *lambda.Lambda, funcArn string, anonAccess bool) error {
	const statementId = "PermitAnonymousAccessToFunctionUrl"
	if anonAccess {
		if err := createAnonUrlAccessPolicy(svc, funcArn, statementId); err != nil {
			return err
		}
	} else {
		_, err := svc.RemovePermission(&lambda.RemovePermissionInput{
			FunctionName: aws.String(funcArn),
			StatementId:  aws.String(statementId),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == lambda.ErrCodeResourceNotFoundException {
					logger.Debugf("anonymous URL access permission did not exist")
					return nil
				} else {
					return fmt.Errorf("failed to delete anonymous URL access permission: %v", err)
				}
			} else {
				return fmt.Errorf("failed to delete anonymous URL access permission: %v", err)
			}
		}
		logger.Debugf("deleted anonymous URL access permission")
	}
	return nil
}

func createAnonUrlAccessPolicy(svc *lambda.Lambda, funcArn string, statementId string) error {
	_, err := svc.AddPermission(&lambda.AddPermissionInput{
		StatementId:         aws.String(statementId),
		Action:              aws.String("lambda:InvokeFunctionUrl"),
		FunctionName:        aws.String(funcArn),
		FunctionUrlAuthType: aws.String(lambda.FunctionUrlAuthTypeNone),
		Principal:           aws.String("*"),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == lambda.ErrCodeResourceConflictException {
				logger.Debugf("anonymous URL access permission already exists")
				return nil
			} else {
				return fmt.Errorf("failed to add anonymous URL access permission: %v", err)
			}
		} else {
			return fmt.Errorf("failed to add anonymous URL access permission: %v", err)
		}
	}
	logger.Debugf("added anonymous URL access permission")
	return nil
}

func (m LambdaRemote) ensureUrlConfigured(svc *lambda.Lambda, funcArn string) (string, error) {
	logger.Debugf("configuring URL for function: %s", funcArn)

	var functionUrl string
	getUrlResult, err := m.checkFunctionUrlConfig(svc, funcArn)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == lambda.ErrCodeResourceNotFoundException {
				urlConfigOutput, err := svc.CreateFunctionUrlConfig(&lambda.CreateFunctionUrlConfigInput{
					AuthType:     aws.String(lambda.FunctionUrlAuthTypeNone),
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
			return "", fmt.Errorf("failed to check if URL config exists for function: %s: %v", funcArn, err)
		}
	} else {
		functionUrl = *getUrlResult.FunctionUrl
		logger.Debugf("function URL already configured: %s", functionUrl)
	}
	return functionUrl, nil
}
