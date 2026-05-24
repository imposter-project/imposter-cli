package engine

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/imposter-project/imposter-cli/internal/prefs"
	"io"
	"net/http"
	"strings"
	"time"
)

const latestReleaseApi = "https://api.github.com/repos/imposter-project/%s/releases/latest"
const checkThresholdSeconds = 86_400

func ResolveLatestToVersion(engineType EngineType, allowCached bool) (string, error) {
	logger.Tracef("resolving latest version (cache allowed: %v)", allowCached)

	now := time.Now().Unix()
	var latest string

	if allowCached {
		latest = loadCached(engineType, now)
	}

	if latest == "" {
		lookup, err := lookupLatest(engineType, now, allowCached)
		if err != nil {
			return "", err
		}
		latest = lookup
	}

	logger.Tracef("resolved latest version: %s", latest)
	return latest, nil
}

// parseMajorVersion returns the major component of the given engine version,
// or (0, false) if the version cannot be parsed as semver. Callers decide how
// to treat the unparseable case (typically: assume the modern 5.x+ line).
func parseMajorVersion(version string) (int64, bool) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return 0, false
	}
	return v.Major, true
}

// UsesEnvConfig reports whether the given engine version expects its config
// directory and listen port via IMPOSTER_CONFIG_DIR and IMPOSTER_PORT env
// vars (5.x and later) rather than --configDir and --listenPort CLI flags
// (4.x and earlier). Unparseable versions (e.g. "dev") default to the env-var
// form.
func UsesEnvConfig(version string) bool {
	major, ok := parseMajorVersion(version)
	if !ok {
		return true
	}
	return major >= 5
}

// DeriveEngineTypeFromVersion returns the engine type implied by an explicit
// engine version, or EngineTypeNone if no derivation can be made. Callers
// should fall back to their configured/default engine type when this returns
// EngineTypeNone.
//
// Versions 5.x and later resolve to EngineTypeNative. Earlier versions, the
// "latest" alias, the empty string, and unparseable values all return
// EngineTypeNone so the default engine continues to apply. (Future: "latest"
// will be re-pointed at v5 and so will yield EngineTypeNative.)
func DeriveEngineTypeFromVersion(version string) EngineType {
	if version == "" || version == "latest" {
		return EngineTypeNone
	}
	major, ok := parseMajorVersion(version)
	if !ok {
		return EngineTypeNone
	}
	if major >= 5 {
		return EngineTypeNative
	}
	return EngineTypeNone
}

func GetHighestVersion(engines []EngineMetadata) string {
	var highest *semver.Version
	for _, engine := range engines {
		v, err := semver.NewVersion(engine.Version)
		if err != nil {
			continue
		}
		if highest == nil || highest.LessThan(*v) {
			highest = v
		}
	}
	if highest != nil {
		return highest.String()
	}
	return ""
}

func loadCached(engineType EngineType, now int64) string {
	var latest string

	p := getVersionPrefs()
	lastCheck, _ := p.ReadPropertyInt(string(engineType) + ".last_version_check")
	if now-int64(lastCheck) < checkThresholdSeconds {
		latest, _ = p.ReadPropertyString(string(engineType) + ".latest")
	}

	logger.Tracef("latest version cached value: %s", latest)
	return latest
}

func lookupLatest(engineType EngineType, now int64, allowFallbackToCached bool) (string, error) {
	apiUrl := fmt.Sprintf(latestReleaseApi, getRepoNameForEngineType(engineType))
	latest, err := fetchLatestFromApi(apiUrl)
	if err != nil {
		if !allowFallbackToCached {
			return "", fmt.Errorf("failed to fetch latest version from API: %s", err)
		}

		logger.Warnf("failed to fetch latest version from API (%s) - checking cache", err)
		latest = loadCached(engineType, now)
		if latest == "" {
			return "", fmt.Errorf("failed to resolve latest version (%s) and no cached version found", err)
		} else {
			// don't persist the cached version back to the prefs store
			return latest, nil
		}
	}

	p := getVersionPrefs()
	err = p.WriteProperty(string(engineType)+".latest", latest)
	if err != nil {
		logger.Warnf("failed to record latest version: %s", err)
	}
	err = p.WriteProperty(string(engineType)+".last_version_check", now)
	if err != nil {
		logger.Warnf("failed to record last version check time: %s", err)
	}
	return latest, nil
}

func getVersionPrefs() prefs.Prefs {
	return prefs.Load("prefs.json")
}

func fetchLatestFromApi(apiUrl string) (string, error) {
	logger.Tracef("fetching latest version from: %s", apiUrl)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("failed to determine latest version from %s: %s", apiUrl, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("failed to determine latest version from %s - status code: %d", apiUrl, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to determine latest version from %s - cannot read response body: %s", apiUrl, err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", fmt.Errorf("failed to determine latest version from %s - cannot unmarshall response body: %s", apiUrl, err)
	}
	tagName := data["tag_name"].(string)
	return strings.TrimPrefix(tagName, "v"), nil
}
