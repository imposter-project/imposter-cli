package engine

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/imposter-project/imposter-cli/internal/prefs"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const latestReleaseApi = "https://api.github.com/repos/imposter-project/%s/releases/latest"
const releaseListApi = "https://api.github.com/repos/imposter-project/%s/releases?per_page=100&page=%d"
const checkThresholdSeconds = 86_400

// maxReleaseListPages bounds how far back through the release list a
// major-version alias lookup will page before giving up.
const maxReleaseListPages = 5

// VersionLatest is the alias meaning "the newest release of the engine".
const VersionLatest = "latest"

func ResolveLatestToVersion(engineType EngineType, allowCached bool) (string, error) {
	logger.Tracef("resolving latest version (cache allowed: %v)", allowCached)

	latest, err := resolveAlias(string(engineType), VersionLatest, allowCached, func() (string, error) {
		return fetchLatestFromApi(fmt.Sprintf(latestReleaseApi, getRepoNameForEngineType(engineType)))
	})
	if err != nil {
		return "", err
	}

	logger.Tracef("resolved latest version: %s", latest)
	return latest, nil
}

// ResolveMajorToVersion resolves a bare major-version alias, such as "4", to
// the highest release carrying that major version, such as "4.9.3". The major
// version alone determines which repository is searched, so this does not
// depend on the configured engine type.
func ResolveMajorToVersion(major int64, allowCached bool) (string, error) {
	logger.Tracef("resolving version alias %d (cache allowed: %v)", major, allowCached)

	repo := getRepoNameForMajor(major)
	resolved, err := resolveAlias(repo, fmt.Sprintf("v%d", major), allowCached, func() (string, error) {
		return fetchHighestForMajor(repo, major)
	})
	if err != nil {
		return "", err
	}

	logger.Tracef("resolved version alias %d: %s", major, resolved)
	return resolved, nil
}

// ParseMajorAlias returns the major version denoted by a bare major-version
// alias, such as "4" or "5", or (0, false) if the given version is not such an
// alias.
func ParseMajorAlias(version string) (int64, bool) {
	major, err := strconv.ParseInt(version, 10, 64)
	if err != nil || major < 0 {
		return 0, false
	}
	return major, true
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

// resolveAlias resolves a version alias, such as "latest" or "v4", to a
// concrete version, using the cached value if permitted and falling back to it
// if the lookup fails. The scope namespaces the cache entry - the engine type
// for "latest", or the repository name for a major version alias.
func resolveAlias(scope string, alias string, allowCached bool, lookup func() (string, error)) (string, error) {
	now := time.Now().Unix()

	if allowCached {
		if cached := loadCached(scope, alias, now); cached != "" {
			return cached, nil
		}
	}

	resolved, err := lookup()
	if err != nil {
		if !allowCached {
			return "", fmt.Errorf("failed to fetch version for alias '%s' from API: %s", alias, err)
		}

		logger.Warnf("failed to fetch version for alias '%s' from API (%s) - checking cache", alias, err)
		cached := loadCached(scope, alias, now)
		if cached == "" {
			return "", fmt.Errorf("failed to resolve version for alias '%s' (%s) and no cached version found", alias, err)
		}
		// don't persist the cached version back to the prefs store
		return cached, nil
	}

	storeCached(scope, alias, resolved, now)
	return resolved, nil
}

// aliasCacheKeys returns the prefs keys holding the resolved version and the
// last check time for the given alias. The "latest" alias keeps its original
// key names so existing prefs files remain valid.
func aliasCacheKeys(scope string, alias string) (versionKey string, checkKey string) {
	if alias == VersionLatest {
		return scope + ".latest", scope + ".last_version_check"
	}
	prefix := scope + "." + alias
	return prefix + ".latest", prefix + ".last_version_check"
}

func loadCached(scope string, alias string, now int64) string {
	var version string

	versionKey, checkKey := aliasCacheKeys(scope, alias)
	p := getVersionPrefs()
	lastCheck, _ := p.ReadPropertyInt(checkKey)
	if now-int64(lastCheck) < checkThresholdSeconds {
		version, _ = p.ReadPropertyString(versionKey)
	}

	logger.Tracef("cached value for alias '%s': %s", alias, version)
	return version
}

func storeCached(scope string, alias string, version string, now int64) {
	versionKey, checkKey := aliasCacheKeys(scope, alias)
	p := getVersionPrefs()
	if err := p.WriteProperty(versionKey, version); err != nil {
		logger.Warnf("failed to record version for alias '%s': %s", alias, err)
	}
	if err := p.WriteProperty(checkKey, now); err != nil {
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

type release struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// fetchHighestForMajor returns the highest released version of the given repo
// carrying the given major version. Draft and pre-release entries are ignored,
// matching the behaviour of the 'latest release' API used for "latest".
func fetchHighestForMajor(repo string, major int64) (string, error) {
	var highest *semver.Version

	for page := 1; page <= maxReleaseListPages; page++ {
		releases, err := fetchReleasesFromApi(fmt.Sprintf(releaseListApi, repo, page))
		if err != nil {
			return "", err
		}
		if len(releases) == 0 {
			break
		}

		pageHighest, sawOlderMajor := selectHighestForMajor(releases, major)
		if pageHighest != nil && (highest == nil || highest.LessThan(*pageHighest)) {
			highest = pageHighest
		}
		// releases are returned newest first, so once older majors appear
		// there is nothing further back worth paging for
		if sawOlderMajor {
			break
		}
	}

	if highest == nil {
		return "", fmt.Errorf("no release found for major version %d of %s", major, repo)
	}
	return highest.String(), nil
}

// selectHighestForMajor returns the highest of the given releases carrying the
// given major version, ignoring drafts, pre-releases and unparseable tags. It
// also reports whether any release with a lower major version was seen.
func selectHighestForMajor(releases []release, major int64) (*semver.Version, bool) {
	var highest *semver.Version
	var sawOlderMajor bool

	for _, r := range releases {
		if r.Draft || r.Prerelease {
			continue
		}
		v, err := semver.NewVersion(strings.TrimPrefix(r.TagName, "v"))
		if err != nil {
			continue
		}
		if v.Major < major {
			sawOlderMajor = true
			continue
		}
		if v.Major > major {
			continue
		}
		if highest == nil || highest.LessThan(*v) {
			highest = v
		}
	}
	return highest, sawOlderMajor
}

func fetchReleasesFromApi(apiUrl string) ([]release, error) {
	logger.Tracef("fetching releases from: %s", apiUrl)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases from %s: %s", apiUrl, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("failed to list releases from %s - status code: %d", apiUrl, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases from %s - cannot read response body: %s", apiUrl, err)
	}
	var releases []release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to list releases from %s - cannot unmarshall response body: %s", apiUrl, err)
	}
	return releases, nil
}
