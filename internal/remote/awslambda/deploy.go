package awslambda

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/imposter-project/imposter-cli/internal/engine/awslambda"
	"github.com/imposter-project/imposter-cli/internal/remote"
	"github.com/imposter-project/imposter-cli/internal/stringutil"
	"os"
	"path"
	"strings"
	"time"
)

const liveAliasName = "live"
const readyTimeoutSeconds = 360

func (m LambdaRemote) Deploy() error {
	region, cfg, svc, err := m.initLambdaClient()
	if err != nil {
		return fmt.Errorf("failed to initialise lambda client: %v", err)
	}

	roleName := stringutil.GetFirstNonEmpty(m.Config[configKeyIamRoleName], defaultIamRoleName)

	roleArn, err := ensureIamRole(cfg, roleName)
	if err != nil {
		logger.Fatal(err)
	}

	engineVersion := engine.GetConfiguredVersion(engine.EngineTypeAwsLambda, m.Config[configKeyEngineVersion], true)
	zipContents, err := awslambda.CreateDeploymentPackage(engineVersion, m.Dir)
	if err != nil {
		logger.Fatal(err)
	}

	var location codeLocation
	if stringutil.ToBoolWithDefault(m.Config[configKeyUploadToS3], true) {
		bucketName, objectKey, err := m.uploadBundleToBucket(zipContents)
		if err != nil {
			return err
		}
		location = codeLocation{
			bucket:    bucketName,
			objectKey: objectKey,
		}
	} else {
		location = codeLocation{
			zipContents: zipContents,
		}
	}

	snapStart := stringutil.ToBool(m.Config[configKeySnapStart])
	funcArn, err := ensureFunctionExists(
		svc,
		region,
		m.getFunctionName(),
		roleArn,
		m.getMemorySize(),
		m.getArchitecture(),
		location,
		snapStart,
	)
	if err != nil {
		return err
	}

	var versionId string
	if stringutil.ToBool(m.Config[configKeyPublishVersion]) {
		versionId, err = publishFunctionVersion(svc, funcArn)
		if err != nil {
			return err
		}
	} else {
		versionId = "$LATEST"
	}

	var arnForUrl string
	if stringutil.ToBool(m.Config[configKeyCreateAlias]) {
		aliasArn, err := createOrUpdateAlias(svc, funcArn, versionId, liveAliasName)
		if err != nil {
			return err
		}
		arnForUrl = aliasArn
	} else {
		arnForUrl = funcArn
	}

	if _, err = m.ensureUrlConfigured(svc, arnForUrl); err != nil {
		return err
	}

	permitAnonAccess := stringutil.ToBool(m.Config[configKeyAnonAccess])
	if err = configureUrlAccess(svc, arnForUrl, permitAnonAccess); err != nil {
		return err
	}
	return nil
}

func ensureSnapStart(svc *lambda.Client, funcArn string, snapStart bool) error {
	var desiredConfig lambdatypes.SnapStartApplyOn
	if snapStart {
		desiredConfig = lambdatypes.SnapStartApplyOnPublishedVersions
	} else {
		desiredConfig = lambdatypes.SnapStartApplyOnNone
	}

	configuration, err := svc.GetFunctionConfiguration(ctx, &lambda.GetFunctionConfigurationInput{FunctionName: aws.String(funcArn)})
	if err != nil {
		return fmt.Errorf("failed to check snapstart configuration for %v: %v", funcArn, err)
	}
	if configuration.SnapStart != nil && configuration.SnapStart.ApplyOn == desiredConfig {
		logger.Tracef("snapstart set to %v for %v", desiredConfig, funcArn)
		return nil
	}

	logger.Tracef("configuring snapstart for %v", funcArn)
	_, err = svc.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(funcArn),
		SnapStart: &lambdatypes.SnapStart{
			ApplyOn: desiredConfig,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to configure snapstart for %v: %v", funcArn, err)
	}
	logger.Tracef("snapstart set to %v for %v", desiredConfig, funcArn)
	return nil
}

func publishFunctionVersion(svc *lambda.Client, funcArn string) (versionId string, err error) {
	// wait for the root function to be ready
	if err = awaitReady(svc, funcArn, ""); err != nil {
		return "", err
	}

	logger.Tracef("publishing version for %v", funcArn)
	version, err := svc.PublishVersion(ctx, &lambda.PublishVersionInput{
		FunctionName: aws.String(funcArn),
	})
	if err != nil {
		return "", err
	}
	versionId = *version.Version

	// wait for the new version to be ready
	if err = awaitReady(svc, funcArn, versionId); err != nil {
		return "", err
	}

	logger.Debugf("published version %v for %v", versionId, funcArn)
	return versionId, nil
}

func awaitReady(svc *lambda.Client, funcArn string, checkVersion string) error {
	logger.Debugf("waiting for function %v [version: %v] to be ready", funcArn, checkVersion)
	for i := 0; i < readyTimeoutSeconds; i++ {
		input := &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(funcArn),
		}
		if checkVersion != "" {
			input.Qualifier = aws.String(checkVersion)
		}
		configuration, err := svc.GetFunctionConfiguration(ctx, input)
		if err != nil {
			return err
		}

		lastUpdateStatus := configuration.LastUpdateStatus
		logger.Tracef("function %v [version: %v] lastUpdateStatus=%v", funcArn, checkVersion, lastUpdateStatus)
		lastUpdateInProgress := lastUpdateStatus != "" && lastUpdateStatus != lambdatypes.LastUpdateStatusSuccessful

		currentState := configuration.State
		logger.Tracef("function %v [version: %v] state=%v", funcArn, checkVersion, currentState)
		stateIsPending := currentState == lambdatypes.StatePending

		if !lastUpdateInProgress && !stateIsPending {
			logger.Debugf("function %v [version: %v] is ready", funcArn, checkVersion)
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timed out after %v seconds waiting for function %v [version: %v] to be ready", readyTimeoutSeconds, checkVersion, funcArn)
}

func createOrUpdateAlias(svc *lambda.Client, funcArn string, versionId string, aliasName string) (aliasArn string, err error) {
	_, err = svc.GetAlias(ctx, &lambda.GetAliasInput{
		FunctionName: aws.String(funcArn),
		Name:         aws.String(aliasName),
	})
	if err != nil {
		var notFoundErr *lambdatypes.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			aliasArn, err := createAlias(svc, funcArn, versionId, aliasName)
			if err != nil {
				return "", fmt.Errorf("failed to create alias: %v", err)
			}
			return aliasArn, nil
		}
		return "", fmt.Errorf("failed to get alias %v for function %v: %v", aliasName, funcArn, err)
	}

	logger.Debugf("updating alias %v for function %v", aliasName, funcArn)
	updateResult, err := svc.UpdateAlias(ctx, &lambda.UpdateAliasInput{
		FunctionName:    aws.String(funcArn),
		FunctionVersion: aws.String(versionId),
		Name:            aws.String(aliasName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to update alias %v for function %v: %v", aliasName, funcArn, err)
	}
	aliasArn = *updateResult.AliasArn
	logger.Debugf("updated alias %v to version %v", aliasArn, versionId)
	return aliasArn, nil
}

func createAlias(svc *lambda.Client, funcArn string, versionId string, aliasName string) (aliasArn string, err error) {
	logger.Tracef("creating alias for function %v to version %v", funcArn, versionId)
	alias, err := svc.CreateAlias(ctx, &lambda.CreateAliasInput{
		FunctionName:    aws.String(funcArn),
		FunctionVersion: aws.String(versionId),
		Name:            aws.String(aliasName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create alias for function %v to version %v: %v", funcArn, versionId, err)
	}
	aliasArn = *alias.AliasArn
	logger.Debugf("created alias %v to version %v", aliasArn, versionId)
	return aliasArn, nil
}

func (m LambdaRemote) Undeploy() error {
	region, _, svc, err := m.initLambdaClient()
	if err != nil {
		return fmt.Errorf("failed to initialise lambda client: %v", err)
	}

	funcName := m.getFunctionName()

	var funcArn string
	funcExists, err := checkFunctionExists(svc, funcName)
	if err != nil {
		var notFoundErr *lambdatypes.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			logger.Debugf("function %s does not exist in region %s", funcName, region)
			return nil
		}
		return fmt.Errorf("failed to check if function %s exists in region %s: %v", funcName, region, err)
	} else {
		funcArn = *funcExists.Configuration.FunctionArn
		logger.Tracef("function ARN: %s", funcArn)
	}

	err = m.deleteFunction(funcArn, svc)
	if err != nil {
		return err
	}
	return nil
}

func (m LambdaRemote) GetEndpoint() (*remote.EndpointDetails, error) {
	_, _, svc, err := m.initLambdaClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialise lambda client: %v", err)
	}

	var funcArn string
	funcExists, err := checkFunctionExists(svc, m.getFunctionName())
	if err != nil {
		return nil, err
	} else {
		funcArn = *funcExists.Configuration.FunctionArn
		logger.Tracef("function ARN: %s", funcArn)
	}

	var functionUrl string
	getUrlResult, err := m.checkFunctionUrlConfig(svc, funcArn)
	if err != nil {
		return nil, err
	} else {
		functionUrl = *getUrlResult.FunctionUrl
		logger.Tracef("function URL: %s", functionUrl)
	}

	details := &remote.EndpointDetails{
		BaseUrl:   functionUrl,
		StatusUrl: remote.MustJoinPath(functionUrl, "/system/status"),

		// spec not supported on lambda
		SpecUrl: "",
	}
	return details, nil
}

func (m LambdaRemote) initLambdaClient() (region string, cfg aws.Config, svc *lambda.Client, err error) {
	if m.Config[configKeyRegion] == "" {
		return "", aws.Config{}, nil, fmt.Errorf("region cannot be null")
	}
	region, cfg, err = m.loadAwsConfig()
	if err != nil {
		return "", aws.Config{}, nil, err
	}
	svc = lambda.NewFromConfig(cfg)
	return region, cfg, svc, nil
}

func (m LambdaRemote) getFunctionName() string {
	configuredFuncName := m.Config[configKeyFuncName]
	if configuredFuncName != "" {
		return configuredFuncName
	}
	funcName := path.Base(m.Dir)
	if !strings.HasPrefix(strings.ToLower(funcName), "imposter") {
		funcName = "imposter-" + funcName
	}
	if len(funcName) > 64 {
		return funcName[:64]
	} else {
		return funcName
	}
}

func (m LambdaRemote) loadAwsConfig() (string, aws.Config, error) {
	region := m.getAwsRegion()
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return "", aws.Config{}, fmt.Errorf("failed to load AWS config: %v", err)
	}
	return region, cfg, nil
}

func ensureFunctionExists(
	svc *lambda.Client,
	region string,
	funcName string,
	roleArn string,
	memoryMb int32,
	arch LambdaArchitecture,
	location codeLocation,
	snapStart bool,
) (string, error) {
	var funcArn string

	result, err := checkFunctionExists(svc, funcName)
	if err != nil {
		var notFoundErr *lambdatypes.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			functionArn, err := createFunction(
				svc,
				region,
				funcName,
				roleArn,
				memoryMb,
				arch,
				location,
				snapStart,
			)
			if err != nil {
				return "", err
			}
			funcArn = functionArn
		} else {
			return "", fmt.Errorf("failed to check if function %s exists in region %s: %v", funcName, region, err)
		}

	} else {
		funcArn = *result.Configuration.FunctionArn
		if err = ensureSnapStart(svc, funcArn, snapStart); err != nil {
			return "", err
		}
		if err = updateFunctionCode(svc, funcArn, location); err != nil {
			return "", err
		}
	}
	return funcArn, nil
}

func checkFunctionExists(svc *lambda.Client, functionName string) (*lambda.GetFunctionOutput, error) {
	result, err := svc.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	})
	return result, err
}

type codeLocation struct {
	bucket      string
	objectKey   string
	zipContents *[]byte
}

func createFunction(
	svc *lambda.Client,
	region string,
	funcName string,
	roleArn string,
	memoryMb int32,
	arch LambdaArchitecture,
	location codeLocation,
	snapStart bool,
) (arn string, err error) {
	logger.Debugf("creating function: %s in region: %s", funcName, region)

	var desiredConfig lambdatypes.SnapStartApplyOn
	if snapStart {
		desiredConfig = lambdatypes.SnapStartApplyOnPublishedVersions
	} else {
		desiredConfig = lambdatypes.SnapStartApplyOnNone
	}

	input := &lambda.CreateFunctionInput{
		FunctionName:  aws.String(funcName),
		Handler:       aws.String("io.gatehill.imposter.awslambda.HandlerV2"),
		MemorySize:    aws.Int32(memoryMb),
		Role:          aws.String(roleArn),
		Runtime:       lambdatypes.RuntimeJava11,
		Architectures: []lambdatypes.Architecture{lambdatypes.Architecture(arch)},
		Environment:   buildEnv(),
		Code:          &lambdatypes.FunctionCode{},
		SnapStart: &lambdatypes.SnapStart{
			ApplyOn: desiredConfig,
		},
	}

	if location.bucket != "" {
		input.Code.S3Bucket = aws.String(location.bucket)
		input.Code.S3Key = aws.String(location.objectKey)
	} else {
		input.Code.ZipFile = *location.zipContents
	}

	result, err := svc.CreateFunction(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create function %s in region %s: %v", funcName, region, err)
	}
	logger.Infof("created function: %s with arn: %s", funcName, *result.FunctionArn)
	return *result.FunctionArn, nil
}

func updateFunctionCode(svc *lambda.Client, funcArn string, location codeLocation) error {
	logger.Debugf("updating function code for: %s", funcArn)
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(funcArn),
	}
	if location.bucket != "" {
		input.S3Bucket = aws.String(location.bucket)
		input.S3Key = aws.String(location.objectKey)
	} else {
		input.ZipFile = *location.zipContents
	}
	_, err := svc.UpdateFunctionCode(ctx, input)
	if err != nil {
		return err
	}
	logger.Infof("updated function code for: %s", funcArn)
	return nil
}

func (m LambdaRemote) checkFunctionUrlConfig(
	svc *lambda.Client,
	funcArn string,
) (*lambda.GetFunctionUrlConfigOutput, error) {

	input := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(funcArn),
	}
	if m.shouldCreateAlias() {
		input.Qualifier = aws.String(m.getFunctionAlias())
	}
	logger.Tracef("checking function URL config for %v", input)
	getUrlResult, err := svc.GetFunctionUrlConfig(ctx, input)
	return getUrlResult, err
}

func (m LambdaRemote) shouldCreateAlias() bool {
	return stringutil.ToBool(m.Config[configKeyCreateAlias])
}

func (m LambdaRemote) getFunctionAlias() string {
	return liveAliasName
}

func (m LambdaRemote) getAwsRegion() string {
	if defaultRegion, ok := os.LookupEnv("AWS_DEFAULT_REGION"); ok {
		return defaultRegion
	} else if configuredRegion := m.Config[configKeyRegion]; configuredRegion != "" {
		return configuredRegion
	}
	panic("no AWS default region set")
}

func buildEnv() *lambdatypes.Environment {
	env := make(map[string]string)
	env["IMPOSTER_CONFIG_DIR"] = "/var/task/config"
	env["JAVA_TOOL_OPTIONS"] = "-XX:+TieredCompilation -XX:TieredStopAtLevel=1"
	return &lambdatypes.Environment{Variables: env}
}

func (m LambdaRemote) deleteFunction(funcArn string, svc *lambda.Client) error {
	logger.Tracef("deleting function: %s", funcArn)
	_, err := svc.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(funcArn),
	})
	if err != nil {
		return fmt.Errorf("failed to delete function: %s: %v", funcArn, err)
	}
	logger.Debugf("deleted function: %s", funcArn)
	return nil
}
