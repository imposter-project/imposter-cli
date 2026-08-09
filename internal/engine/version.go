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

// VersionLatest is the alias meaning "the newest release of the engine".
const VersionLatest = "latest"

// majorAliases are the supported short version aliases, each meaning "the
// newest release of that engine line": "4" is the JVM engine and "5" is the
// native engine.
var majorAliases = map[string]int64{
	"4": 4,
	"5": 5,
}

func ResolveLatestToVersion(engineType EngineType, allowCached bool) (string, error) {
	logger.Tracef("resolving latest version (cache allowed: %v)", allowCached)

	latest, err := resolveLatest(string(engineType), allowCached, func() (string, error) {
		return fetchLatestFromApi(fmt.Sprintf(latestReleaseApi, getRepoNameForEngineType(engineType)))
	})
	if err != nil {
		return "", err
	}

	logger.Tracef("resolved latest version: %s", latest)
	return latest, nil
}

// ResolveMajorToVersion resolves a major-version alias to the latest release
// of the corresponding engine line, so "4" resolves to the latest JVM engine
// release and "5" to the latest native engine release. The major version alone
// determines which repository is used, so this does not depend on the
// configured engine type.
func ResolveMajorToVersion(major int64, allowCached bool) (string, error) {
	logger.Tracef("resolving version alias %d (cache allowed: %v)", major, allowCached)

	repo := getRepoNameForMajor(major)
	resolved, err := resolveLatest(repo, allowCached, func() (string, error) {
		return fetchLatestFromApi(fmt.Sprintf(latestReleaseApi, repo))
	})
	if err != nil {
		return "", err
	}

	logger.Tracef("resolved version alias %d: %s", major, resolved)
	return resolved, nil
}

// ParseMajorAlias returns the major version denoted by a major-version alias,
// or (0, false) if the given version is not one of the supported aliases.
func ParseMajorAlias(version string) (int64, bool) {
	major, ok := majorAliases[version]
	return major, ok
}

// parseMajorVersion returns the major component of the given engine version,
// accepting either a full semver version or a bare major-version alias such as
// "4". It returns (0, false) if the version can be parsed as neither. Callers
// decide how to treat the unparseable case (typically: assume the modern 5.x+
// line).
func parseMajorVersion(version string) (int64, bool) {
	if major, ok := ParseMajorAlias(version); ok {
		return major, true
	}
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
// Versions 5.x and later resolve to EngineTypeNative, including the bare "5"
// major-version alias. Earlier versions, the "latest" alias, the empty string,
// and unparseable values all return EngineTypeNone so the default engine
// continues to apply. (Future: "latest" will be re-pointed at v5 and so will
// yield EngineTypeNative.)
func DeriveEngineTypeFromVersion(version string) EngineType {
	if version == "" || version == VersionLatest {
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
	return getHighestVersion(engines, nil)
}

// GetHighestVersionForMajor returns the highest of the given engine versions
// carrying the given major version, or the empty string if none do.
func GetHighestVersionForMajor(engines []EngineMetadata, major int64) string {
	return getHighestVersion(engines, &major)
}

func getHighestVersion(engines []EngineMetadata, major *int64) string {
	var highest *semver.Version
	for _, engine := range engines {
		v, err := semver.NewVersion(engine.Version)
		if err != nil {
			continue
		}
		if major != nil && v.Major != *major {
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

// resolveLatest resolves the latest version for the given scope, using the
// cached value if permitted and falling back to it if the lookup fails. The
// scope namespaces the cache entry - the engine type when resolving "latest",
// or the repository name when resolving a major version alias.
func resolveLatest(scope string, allowCached bool, lookup func() (string, error)) (string, error) {
	now := time.Now().Unix()

	if allowCached {
		if cached := loadCached(scope, now); cached != "" {
			return cached, nil
		}
	}

	resolved, err := lookup()
	if err != nil {
		if !allowCached {
			return "", fmt.Errorf("failed to fetch latest version from API: %s", err)
		}

		logger.Warnf("failed to fetch latest version from API (%s) - checking cache", err)
		cached := loadCached(scope, now)
		if cached == "" {
			return "", fmt.Errorf("failed to resolve latest version (%s) and no cached version found", err)
		}
		// don't persist the cached version back to the prefs store
		return cached, nil
	}

	storeCached(scope, resolved, now)
	return resolved, nil
}

func loadCached(scope string, now int64) string {
	var version string

	p := getVersionPrefs()
	lastCheck, _ := p.ReadPropertyInt(scope + ".last_version_check")
	if now-int64(lastCheck) < checkThresholdSeconds {
		version, _ = p.ReadPropertyString(scope + ".latest")
	}

	logger.Tracef("latest version cached value for %s: %s", scope, version)
	return version
}

func storeCached(scope string, version string, now int64) {
	p := getVersionPrefs()
	if err := p.WriteProperty(scope+".latest", version); err != nil {
		logger.Warnf("failed to record latest version: %s", err)
	}
	if err := p.WriteProperty(scope+".last_version_check", now); err != nil {
		logger.Warnf("failed to record last version check time: %s", err)
	}
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
