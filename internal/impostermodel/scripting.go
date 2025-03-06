/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package impostermodel

import (
	"fmt"
	"gatehill.io/imposter/internal/fileutil"
	"os"
	"path/filepath"
)

type ScriptEngine string

const (
	ScriptEngineNone       ScriptEngine = "none"
	ScriptEngineGroovy     ScriptEngine = "groovy"
	ScriptEngineJavaScript ScriptEngine = "javascript"
)

// shorthand for ScriptEngineJavaScript
const jsScriptEngineAlias = "js"

func ParseScriptEngine(scriptEngine string) ScriptEngine {
	engine := ScriptEngine(scriptEngine)
	switch engine {
	case ScriptEngineNone, ScriptEngineGroovy, ScriptEngineJavaScript:
		return engine
	case jsScriptEngineAlias:
		return ScriptEngineJavaScript
	case "":
		return ScriptEngineNone
	default:
		panic(fmt.Errorf("unsupported script engine: %v", scriptEngine))
	}
}

func BuildScriptFilePath(anchorFilePath string, scriptEngine ScriptEngine, forceOverwrite bool) string {
	var scriptFilePath string
	if IsScriptEngineEnabled(scriptEngine) {
		var scriptEngineExt string
		switch scriptEngine {
		case ScriptEngineJavaScript:
			scriptEngineExt = ".js"
			break
		case ScriptEngineGroovy:
			scriptEngineExt = ".groovy"
			break
		default:
			logger.Fatal("script engine is disabled")
		}
		scriptFilePath = fileutil.GenerateFilePathAdjacentToFile(anchorFilePath, scriptEngineExt, forceOverwrite)
	}
	return scriptFilePath
}

func IsScriptEngineEnabled(engine ScriptEngine) bool {
	return len(engine) > 0 && engine != ScriptEngineNone
}

func getScriptFileName(anchorFilePath string, scriptEngine ScriptEngine, forceOverwrite bool) string {
	var scriptFileName string
	if IsScriptEngineEnabled(scriptEngine) {
		scriptFilePath := writeScriptFile(anchorFilePath, scriptEngine, forceOverwrite)
		scriptFileName = filepath.Base(scriptFilePath)
	}
	return scriptFileName
}

func writeScriptFile(anchorFilePath string, engine ScriptEngine, forceOverwrite bool) string {
	scriptFilePath := BuildScriptFilePath(anchorFilePath, engine, forceOverwrite)
	scriptFile, err := os.Create(scriptFilePath)
	if err != nil {
		logger.Fatalf("error writing script file: %v: %v", scriptFilePath, err)
	}
	defer scriptFile.Close()

	_, err = scriptFile.WriteString(`
// TODO add your custom logic here
logger.debug('method: ' + context.request.method);
logger.debug('path: ' + context.request.path);
logger.debug('pathParams: ' + context.request.pathParams);
logger.debug('queryParams: ' + context.request.queryParams);
logger.debug('headers: ' + context.request.headers);
`)
	if err != nil {
		logger.Fatalf("error writing script file: %v: %v", scriptFilePath, err)
	}

	logger.Infof("wrote script file: %v", scriptFilePath)
	return scriptFilePath
}
