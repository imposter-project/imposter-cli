package awslambda

import (
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/imposter-project/imposter-cli/internal/engine"
)

// lambdaFlavour captures the runtime-specific differences between the
// JVM-engine AWS Lambda binary (imposter-jvm-engine) and the native
// imposter-go binary that runs on a custom Amazon Linux runtime.
type lambdaFlavour struct {
	runtime           lambdatypes.Runtime
	handler           string
	envVars           map[string]string
	supportsSnapStart bool
}

var jvmLambdaFlavour = lambdaFlavour{
	runtime: lambdatypes.RuntimeJava11,
	handler: "io.gatehill.imposter.awslambda.HandlerV2",
	envVars: map[string]string{
		"IMPOSTER_CONFIG_DIR": "/var/task/config",
		"JAVA_TOOL_OPTIONS":   "-XX:+TieredCompilation -XX:TieredStopAtLevel=1",
	},
	supportsSnapStart: true,
}

// nativeLambdaFlavour targets the imposter-go binary deployed on the
// provided.al2023 custom runtime. The handler must be named "bootstrap" by
// AWS convention (the zip's executable is invoked directly). SnapStart is a
// JVM-only feature and is not configured for this flavour.
var nativeLambdaFlavour = lambdaFlavour{
	runtime: lambdatypes.RuntimeProvidedal2023,
	handler: "bootstrap",
	envVars: map[string]string{
		"IMPOSTER_CONFIG_DIR": "/var/task/config",
	},
	supportsSnapStart: false,
}

// flavourForVersion picks the Lambda flavour implied by the engine version,
// matching the binary-download gating in engine/awslambda/binary.go.
func flavourForVersion(version string) lambdaFlavour {
	if engine.DeriveEngineTypeFromVersion(version) == engine.EngineTypeNative {
		return nativeLambdaFlavour
	}
	return jvmLambdaFlavour
}

func (f lambdaFlavour) buildEnv() *lambdatypes.Environment {
	env := make(map[string]string, len(f.envVars))
	for k, v := range f.envVars {
		env[k] = v
	}
	return &lambdatypes.Environment{Variables: env}
}
