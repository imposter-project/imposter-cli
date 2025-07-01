# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.3.0] - 2025-07-01
### Added
- feat: add plugin uninstall command.
- feat: version lookup depends on engine type.

### Changed
- refactor: remove redundant platform package
- refactor: use templates for plugin file name config.

### Fixed
- fix: escape metacharacters in plugin file pattern.

## [1.2.1] - 2025-06-30
### Changed
- refactor: make full mode optional in version command.

## [1.2.0] - 2025-06-30
### Added
- feat: adds golang to engine type option description.
- feat: version command can output just the CLI version.

## [1.1.6] - 2025-06-29
### Changed
- chore(deps): bump github.com/docker/docker
- chore(deps): bump github.com/shirou/gopsutil/v4 from 4.25.4 to 4.25.5
- test: bumps imposter-go version.

### Fixed
- fix: temp pattern should handle file extensions with multiple parts.

## [1.1.5] - 2025-06-29
### Changed
- build: adds since config.
- chore: ignore tool config.
- chore: update Go version to 1.23.8.
- refactor(golang): standardise environment creation and add plugin support.
- refactor: aligns go binary naming to expected platform format.
- refactor: maps plugin config to engine type.
- refactor: moves plugin filename prefix to const.
- refactor: plugin downloads are engine type aware.
- refactor: support extracting both zip and tar.gz archives.
- refactor: use a temp file when downloading binaries.

## [1.1.4] - 2025-05-15
### Changed
- chore(deps): bump github.com/aws/aws-sdk-go from 1.55.6 to 1.55.7
- chore(deps): bump github.com/shirou/gopsutil/v4 from 4.25.3 to 4.25.4

### Fixed
- fix(docker): don't prefix docker.io when a custom registry is set.

## [1.1.3] - 2025-04-24
### Changed
- chore(deps): bump github.com/docker/docker

### Fixed
- fix: update Docker API references for v28.1.1 compatibility

## [1.1.2] - 2025-04-24
### Changed
- chore(deps): bump github.com/shirou/gopsutil/v4 from 4.25.1 to 4.25.3
- chore(deps): bump github.com/spf13/cobra from 1.8.1 to 1.9.1
- docs: updates brew tap repo.

### Fixed
- fix: JVM engine environment should include path.

## [1.1.1] - 2025-04-12
### Changed
- ci: updates brew formula repo.

### Fixed
- fix: update format_overrides to use formats array in goreleaser config
- fix: update goreleaser config to fix deprecation warnings
- fix: update goreleaser config to v2 standards

## [1.1.0] - 2025-03-30
### Added
- feat: allow container registry to be set. (#40)

### Changed
- build: ignore dev guide.

## [1.0.6] - 2025-03-28
### Fixed
- fix: mirror upstream status code in client response.

## [1.0.5] - 2025-03-10
### Changed
- refactor: updates GitHub org to imposter-project.

## [1.0.4] - 2025-03-06
### Changed
- ci: splits release into separate job.
- refactor: moves packages to internal.

## [1.0.3] - 2025-02-03
### Changed
- chore(deps): bump github.com/shirou/gopsutil/v4 from 4.24.12 to 4.25.1
- ci: reworks release binary names and download script.

## [1.0.2] - 2025-01-30
### Fixed
- fix: golang engine start should await healthcheck.

## [1.0.1] - 2025-01-30
### Changed
- test: improves coverage of OpenAPI discovery.
- test: improves coverage of OpenAPI parsing.
- test: improves coverage of fileutil.
- test: improves coverage of library.
- test: improves coverage of string utils.

## [1.0.0] - 2025-01-30
### Added
- build: first stable release.

## [0.46.2] - 2025-01-30
### Changed
- chore(deps): bump github.com/aws/aws-sdk-go from 1.55.5 to 1.55.6
- chore(deps): bump github.com/google/uuid from 1.4.0 to 1.6.0

## [0.46.1] - 2025-01-29
### Changed
- build: updates org name.

## [0.46.0] - 2025-01-06
### Added
- feat: adds golang engine.
- feat: golang engine should use binary cache.
- feat: golang engine supports list and stop.

### Changed
- chore: bumps gopsutil to v4.
- refactor: aligns golang binary naming.

### Fixed
- fix: searches golang engine environment for port.

## [0.45.1] - 2025-01-03
### Changed
- build: ignore tool versions config.
- chore(deps): bump github.com/spf13/cobra from 1.4.0 to 1.8.1
- chore(deps): bump github.com/spf13/viper from 1.10.0 to 1.19.0
- chore(deps): bump goreleaser/goreleaser-action from 5 to 6
- refactor: place jvm binaries under their own subdirectory.

## [0.45.0] - 2024-11-20
### Added
- feat: proxy can capture request body and headers.

### Changed
- chore(deps): bump github.com/shirou/gopsutil/v3 from 3.22.2 to 3.24.5
- chore(deps): bump golang.org/x/mod from 0.8.0 to 0.22.0
- chore: bumps engine test version to 4.2.2.
- docs: describes request body and header capture flags.
- refactor: removes dependency on golang.org/x/mod package for semver comparison.
- test: improves coverage of media type detector.
- test: improves coverage of proxy recorder.

## [0.44.1] - 2024-11-17
### Changed
- build: adds dependabot config.
- build: sets package ecosystem to gomod.
- chore(deps): bump github.com/docker/go-connections from 0.4.0 to 0.5.0
- chore(deps): bump github.com/stretchr/testify from 1.7.0 to 1.9.0
- chore(deps): bumps golang to 1.23.3.

## [0.44.0] - 2024-08-02
### Added
- feat(awslambda): allows S3 object key to be set.
- feat(awslambda): awaits version publish completion.
- feat(awslambda): supports bucket upload.
- feat(awslambda): supports publishing version, creating alias and enabling snapstart.

## [0.43.0] - 2024-07-02
### Added
- feat: adds support for docker-distroless image.

### Changed
- chore: bumps engine test version to 3.44.1.

## [0.42.3] - 2024-06-09
### Changed
- build(deps): bump github.com/docker/docker

## [0.42.2] - 2024-06-09
### Changed
- chore: bumps golang to 1.22.4.

## [0.42.1] - 2024-06-06
### Changed
- docs: describes debug mode flag.
- refactor: improves logging if plugins or file cache are enabled.

## [0.42.0] - 2024-05-01
### Added
- feat: adds support for starting in debug mode.

## [0.41.3] - 2024-04-29
### Fixed
- fix: default plugin install tip should only display if flag is false.

## [0.41.2] - 2024-04-22
### Changed
- chore(deps): bumps actions/checkout to v4 and actions/setup-go to v4.
- chore(deps): bumps actions/setup-go to v5.
- chore(deps): bumps actions/setup-java to v4.
- chore(deps): bumps go to 1.21.
- chore(deps): bumps goreleaser/goreleaser-action to v5.
- ci: replaces deprecated Goreleaser config.
- test: bumps engine to 3.38.2.

## [0.41.1] - 2024-04-22
### Changed
- refactor: adds init as alias for scaffold command.
- refactor: improves REST and OpenAPI scaffolding.

## [0.41.0] - 2024-04-22
### Added
- feat: allows minimum CLI version to be specified.

## [0.40.0] - 2024-04-22
### Added
- feat: allows required plugins to be set in local CLI config file.
- feat: supports specification of environment variables in config files.

### Changed
- build(deps): bump golang.org/x/net from 0.18.0 to 0.23.0

## [0.39.0] - 2024-04-15
### Added
- feat: allow plugins to be specified in CLI config.

### Changed
- docs: adds changelog.

### Fixed
- fix: don't quieten the default logger if trace is enabled.

## [0.38.0] - 2024-03-08
### Added
- feat: allows lambda binary path to be configured.

## [0.37.0] - 2024-02-11
### Added
- feat: adds support for archive format plugins.
- feat: adds support for linux/aarch64 to install script.

### Changed
- test: improves coverage of Docker bundle builder.

## [0.36.0] - 2023-12-20
### Added
- feat: adds bundle support for Docker engine.
- feat: improves default docker bundle image name.

### Changed
- refactor: removes unneeded method from function signature.

## [0.35.3] - 2023-11-18
### Changed
- build: reinstates dependency updates.
- ci: pins Java version to last successful.

## [0.35.2] - 2023-11-18
### Changed
- build: reverts module updates to address test regression.
- ci: bumps runner size.
- ci: sets workflow timeout.
- test: bumps test engine version.
- test: disabled flaky test.

## [0.35.1] - 2023-11-18
### Changed
- build(deps): bump golang.org/x/net from 0.7.0 to 0.17.0
- build(deps): bump gopkg.in/yaml.v3
- chore(deps): refresh dependencies.

## [0.35.0] - 2023-11-18
### Added
- feat: adds local config writer command.

## [0.34.1] - 2023-10-26
### Changed
- build(deps): bump github.com/docker/distribution
- test: improves coverage for command line argument parser.

## [0.34.0] - 2023-10-26
### Added
- feat(jvm): fall back to default listen port when listing mocks, if unspecified.

## [0.33.0] - 2023-09-04
### Added
- feat: adds awslambda to bundle command engine type flag completions.
- feat: adds quiet option to list command.

### Changed
- docs: improves explanation of list command.

## [0.32.2] - 2023-07-16
### Fixed
- fix: removes goreleaser replacements config.

## [0.32.1] - 2023-04-26
### Changed
- refactor: makes exclusion of hidden files when listing files explicit.

## [0.32.0] - 2023-04-26
### Added
- feat: adds bundle command.
- feat: adds log level flag completions.

### Changed
- refactor: moves config validation to subpackage.

## [0.31.1] - 2023-04-23
### Changed
- build(deps): bump github.com/docker/docker

## [0.31.0] - 2023-04-23
### Added
- feat: adds command to undeploy from remote.

### Changed
- chore(deps): bumps aws-sdk-go to v1.44.248.

## [0.30.0] - 2023-04-22
### Added
- feat: allows lambda architecture to be set.

### Changed
- docs: adds upgrade instructions.

## [0.29.0] - 2023-04-14
### Added
- feat: read CLI config file from mock config dir.

## [0.28.0] - 2023-03-16
### Added
- feat: setting CLI log level also sets engine log level.

### Other
- Merge pull request #4 from gatehill/dependabot/go_modules/golang.org/x/net-0.7.0

## [0.27.1] - 2023-02-07
### Fixed
- fix: enables default plugin retrieval to differ based on engine type.

## [0.27.0] - 2023-02-07
### Added
- feat: adds shell completions for engine type flag.

## [0.26.0] - 2023-02-07
### Added
- feat: adds 'all' distro engine type.

## [0.25.0] - 2023-01-15
### Added
- feat: splits remote endpoint retrieval from deployment.

## [0.24.3] - 2023-01-08
### Fixed
- fix(awslambda): reports AWS error details.

## [0.24.2] - 2023-01-08
### Changed
- refactor: separates command to set remote type.

### Fixed
- fix(awslambda): max function name check should allow shorter values.

## [0.24.1] - 2023-01-02
### Fixed
- fix: added warning when installing plugins as non-defaults.

## [0.24.0] - 2023-01-02
### Added
- feat: enables plugin list command to enumerate all installed versions.

## [0.23.0] - 2023-01-02
### Added
- feat: adds plugin list command.

## [0.22.2] - 2023-01-01
### Fixed
- fix(proxy): empty response body should skip writing a response file.

## [0.22.1] - 2022-09-09
### Changed
- docs: improves readme.

### Fixed
- fix: improves uniqueness of proxy response file names.

## [0.22.0] - 2022-09-08
### Added
- feat: infers status code when scaffolding from OpenAPI.

## [0.21.0] - 2022-08-23
### Added
- feat: adds suggested workspace names as shell completions.
- feat: improves relevance of doctor hints.
- feat: shows remote provider type in workspace show output.

### Fixed
- fix: removes leading and trailing quotes when setting remote config.

## [0.20.0] - 2022-08-23
### Added
- feat: adds AWS Lambda deployment.
- feat: adds alias for workspace command.
- feat: adds workspace show command.
- feat: allows engine version to be specified in lambda deployments.
- feat: allows lambda memory and URL access to be configured.
- feat: improves remote configuration.

### Changed
- docs: improves scaffold description.

### Fixed
- fix: tolerate empty token when showing remote.

## [0.19.0] - 2022-08-22
### Added
- feat: adds sealed distro library type.

### Fixed
- fix: prevents resolution of latest version in version command.
- fix: removes unnecessary arg from rewrite debug log.

## [0.18.1] - 2022-08-22
### Added
- feat: version command uses cached engines instead of pulling.

## [0.18.0] - 2022-08-21
### Added
- feat: adds URL rewrite to proxy.
- feat: adds hierarchical response file structure to proxy.
- feat: adds option to ignore recording of duplicate requests in proxy.
- feat: adds option to record only certain response headers.
- feat: adds proxy recording.
- feat: adds reverse proxy.
- feat: adds shorthand -c flag for cli-only in version command.
- feat: allows proxy transport to be configured.
- feat: deduplicates response files in proxy.
- feat: generates single config file for proxies.
- feat: supports additional JS media types for proxy URL rewrite.
- feat: switches to impostermodel for config generation for proxies.

### Changed
- build: bumps golang to 1.19.
- docs: adds missing copyright headers.
- docs: describes proxy command.
- test: improves coverage for proxy recorder.

### Fixed
- fix: passes through query string to proxy upstream.
- fix: removes early exit in install script.

## [0.17.0] - 2022-08-09
### Added
- feat: adds ls alias for engine list and workspace list commands.
- feat: adds rest mock scaffolding.

## [0.16.2] - 2022-08-07
### Added
- feat: adds JSON output format and CLI-only mode to version command.

## [0.16.1] - 2022-08-01
### Fixed
- fix: detects unpacked JVM mock processes.

## [0.16.0] - 2022-08-01
### Added
- feat: adds argument to enable recursive config scan in up command.
- feat: adds ls alias for mock list command.
- feat: detects listen port for JVM mocks.

### Fixed
- fix: makes java process detection more robust.

## [0.15.1] - 2022-07-29
### Added
- feat: allows list command exit code to be set based on mock health.

## [0.15.0] - 2022-07-29
### Added
- feat: adds port and health to mock list.

## [0.14.2] - 2022-07-29
### Fixed
- fix: quietens mock list logs.

## [0.14.1] - 2022-07-29
### Fixed
- fix: removes unnecessary client close.

## [0.14.0] - 2022-07-28
### Added
- feat: adds mock list command.
- feat: adds support for ARM64 on macOS to install script.

### Changed
- docs: updates JVM version.

## [0.13.0] - 2022-07-28
### Added
- feat: moves to dedicated logger instance in place of default.

### Changed
- chore: bumps docker to 20.10.17.
- test: bumps distro version to 3.0.2.
- test: finds free port for engine tests.

## [0.12.7] - 2022-07-28
### Changed
- refactor: moves engine build into provider.

## [0.12.6] - 2022-03-27
### Changed
- chore: bumps dependencies for docker, gopsutil and cobra.
- ci: uses go 1.17 compatibility for module pruning.

## [0.12.5] - 2022-03-26
### Fixed
- fix: deduplicates environment variables when provided and present in parent.

## [0.12.4] - 2022-03-23
### Changed
- refactor: logs engine type at trace level.

## [0.12.3] - 2022-02-22
### Changed
- refactor: switches lifecycle state from active to live.

## [0.12.2] - 2022-02-22
### Added
- feat: adds remote status command.

## [0.12.1] - 2022-02-22
### Changed
- refactor: shortens workspace metadata dir name.

## [0.12.0] - 2022-02-21
### Added
- feat: adds cloudmocks remote deployment.
- feat: adds workspaces.
- feat: allows workspace path to be specified.

### Changed
- docs: adds plugin install save default.
- refactor: allows prefs file name to be set.
- refactor: renames meta package to prefs.

### Fixed
- fix: checks if path is a dir when ensuring it exists.

## [0.11.8] - 2022-02-11
### Added
- feat: allows installed plugins to be enabled by default.

### Changed
- docs: describes log level argument.

## [0.11.7] - 2022-01-31
### Added
- feat: adds short form directory mount format.
- feat: allows log level to be set using argument.
- feat: validates directory mounts before binding.

## [0.11.6] - 2022-01-31
### Added
- feat: allows additional bind-mounts to be passed to Docker engine type.

## [0.11.5] - 2022-01-28
### Added
- feat: improves healthcheck failure logging.

## [0.11.4] - 2022-01-27
### Fixed
- fix: debounces dir watcher over longer duration.

## [0.11.3] - 2022-01-27
### Added
- feat: installs missing default plugins on engine start.

## [0.11.2] - 2022-01-27
### Added
- feat: adds support for recursive configuration file scanning.
- feat: injects explicit environment variables into config scope.

### Changed
- docs: adds plugins to configuration file reference.

## [0.11.1] - 2022-01-27
### Fixed
- fix: work-around for comma-separated plugins in environment variable.

## [0.11.0] - 2022-01-27
### Added
- feat: allows multiple plugins to be installed.

### Changed
- ci: pins Java version for jvm engine tests.
- docs: updates plugin install usage.
- test: bumps engine version.
- test: improves coverage for ensuring plugin dir and file cache dir.

### Fixed
- fix: removes empty local library file on download failure.

## [0.10.1] - 2022-01-16
### Added
- feat: adds support for remote file cache.

### Changed
- refactor: moves fallback engine filename to calling function.
- refactor: renames engine binary dir.
- test: improves cleanup for meta store.
- test: improves coverage for engine build util functions.
- test: improves coverage for engine builder.
- test: improves coverage for ensuring plugin dir.

## [0.10.0] - 2022-01-12
### Added
- feat: adds meta store.
- feat: resolves 'latest' version string to actual version via cached lookup.

### Fixed
- fix: passes HOME environment variable to JVM engine.

## [0.9.5] - 2022-01-12
### Changed
- refactor: changes Java discovery to prefer PATH before JAVA_HOME.

## [0.9.4] - 2022-01-12
### Added
- feat: uses IMPOSTER_PLUGIN_DIR as plugin store if set.

## [0.9.3] - 2022-01-12
### Changed
- refactor: improves plugin base directory config key.

## [0.9.2] - 2022-01-12
### Changed
- refactor: moves providers inside library.

## [0.9.1] - 2022-01-10
### Added
- feat: allows explicit environment variables to be passed to engine.

## [0.9.0] - 2022-01-10
### Added
- feat: adds plugin install command.

### Changed
- ci: switches dependency fetch to use mod download.
- refactor: moves cache management and binary download to separate package.

### Fixed
- fix: shows usage if root command invoked with no arguments.

## [0.8.3] - 2022-01-09
### Added
- feat: adds engine list command.

## [0.8.2] - 2022-01-05
### Added
- feat: enables container user to be set.

### Changed
- chore: bumps engine version to 2.4.13.

## [0.8.1] - 2022-01-05
### Changed
- chore: bumps engine version to 2.4.12.
- refactor: switches configuration directory from mount to bind.

## [0.8.0] - 2021-12-21
### Added
- feat: unpacked engine now uses direct java command invocation with classpath.

## [0.7.12] - 2021-12-21
### Changed
- refactor: supports JAVA_TOOL_OPTIONS over JAVA_OPTS environment variable.

## [0.7.11] - 2021-12-21
### Changed
- refactor: CLI version falls back to dev if unset.

## [0.7.10] - 2021-12-19
### Changed
- refactor: returns outcome when starting engine.

## [0.7.9] - 2021-12-19
### Added
- feat: stops healthcheck wait on sigterm.

### Changed
- chore: bumps dependencies.

## [0.7.8] - 2021-12-19
### Added
- feat: improves API when used as an embedded library.

## [0.7.7] - 2021-12-17
### Added
- feat: adds support for version flag in root command.

## [0.7.6] - 2021-12-15
### Fixed
- fix: improves locking of debouncer IDs.

## [0.7.5] - 2021-12-15
### Added
- feat: passes through JAVA_OPTS environment variable.

## [0.7.4] - 2021-12-15
### Added
- feat: adds support for unpacked distributions.
- feat: discovers JAVA_HOME when using unpacked engine type.

## [0.7.3] - 2021-11-23
### Changed
- refactor: bumps log level of dir watch event.

## [0.7.2] - 2021-11-23
### Added
- feat: allows JVM engine binary cache and JAR path to be specified.
- feat: passes through Imposter environment variables to engine.

### Changed
- build: use posix uname flag to determine architecture.
- docs: updates homepage link.
- docs: updates links to documentation.

### Other
- tests: bumps engine version.

## [0.7.1] - 2021-10-26
### Changed
- test: improves coverage of JVM engine pull.

## [0.7.0] - 2021-10-26
### Added
- feat: allow use of 'latest' tag with JVM engine.

### Changed
- docs: updates usage for pull

## [0.6.19] - 2021-10-06
### Changed
- refactor: return engine errors instead of panic

## [0.6.18] - 2021-10-06
### Changed
- refactor: determines version consistently

## [0.6.17] - 2021-10-06
### Added
- feat: gets version output from engine, if possible

## [0.6.16] - 2021-10-06
### Added
- feat: adds JVM support to 'down' command

### Changed
- ci: adds release dry run if revision is not tagged

### Fixed
- fix: bumps go-ole dep for windows_arm64 support

## [0.6.15] - 2021-10-05
### Added
- feat: adds 'down' command with Docker engine support.

## [0.6.14] - 2021-10-01
### Changed
- chore: bumps fallback version.
- docs: updates pull documentation

## [0.6.13] - 2021-09-30
### Changed
- refactor: simplifies config fallback.

## [0.6.12] - 2021-09-29
### Fixed
- fix: restores default engine type of Docker.

## [0.6.11] - 2021-09-29
### Other
- Adds engine pull command.

## [0.6.10] - 2021-09-28
### Other
- Catches binary download errors more reliably.

## [0.6.9] - 2021-09-27
### Other
- Improves deduplication logging.

## [0.6.8] - 2021-09-27
### Other
- Enables customisation of deduplication hash.

## [0.6.7] - 2021-09-27
### Other
- Removes unused block until stopped function.
- Removes unused notify on stop function.

## [0.6.5] - 2021-09-24
### Other
- Updates doctor output with engine type instructions.

## [0.6.4] - 2021-09-24
### Other
- Replaces running containers with identical port and config directory.

## [0.6.3] - 2021-09-23
### Other
- Improves Docker doctor output.
- Improves up scaffold flag description.
- Switches to better test assertions.

## [0.6.2] - 2021-09-23
### Other
- Adds doctor command.
- Adds documentation about different mock engines.

## [0.6.1] - 2021-09-22
### Other
- Improves error message for invalid engine type.

## [0.6.0] - 2021-09-21
### Other
- Enables automatic scaffolding of missing configuration.
- Use release token.

## [0.5.7] - 2021-09-21
### Other
- Adds CI workflow.
- Adds Java binary file extension on Windows.
- Allows CLI configuration file to be specified.
- Start now waits for healthy engine.

## [0.5.6] - 2021-09-17
### Other
- Sets environment variable prefix.

## [0.5.5] - 2021-09-17
### Other
- Bumps fallback version.
- Enables engine version to be set using configuration file.
- Improves test coverage for engines, scaffold and version commands.

## [0.5.4] - 2021-09-14
### Other
- Enables engine type to be set using configuration file.

## [0.5.2] - 2021-09-14
### Other
- Moves event dispatch to separate package.

## [0.5.1] - 2021-09-14
### Other
- Enables auto-restart in JVM engine.

## [0.5.0] - 2021-09-13
### Other
- Adds JVM engine implementation.

## [0.4.5] - 2021-09-12
### Other
- Adds JavaScript script engine argument shorthand.

## [0.4.4] - 2021-09-12
### Other
- Makes script path generation absolute.

## [0.4.3] - 2021-09-11
### Other
- Separates resource generation from mock config generation.
- Simplifies scaffold file path generation.

## [0.4.2] - 2021-09-10
### Other
- Abstracts engine implementation behind interface.

## [0.4.1] - 2021-09-10
### Other
- Adds shorthand for overwrite and script engine args.

## [0.4.0] - 2021-09-09
### Other
- Adds license headers.
- Automatically restart if config directory contents change.
- Updates documentation.

## [0.3.12] - 2021-09-06
### Other
- Generates Imposter resources based on OpenAPI specification.

## [0.3.10] - 2021-09-03
### Other
- Moves engine start options to struct and allows log level to be customised.

## [0.3.9] - 2021-09-02
### Other
- Moves engine management functions to separate package.

## [0.3.8] - 2021-09-02
### Other
- Adds version subcommand.

## [0.3.7] - 2021-09-02
### Other
- Adds scaffold command.
- Ensures config directory path is absolute.

## [0.3.6] - 2021-08-31
### Other
- Allows image download to be forced.
- Bumps Go to 1.17.

## [0.3.4] - 2021-08-30
### Other
- Adds Homebrew instructions.

## [0.3.2] - 2021-08-30
### Other
- Adds Homebrew tap configuration.

## [0.3.1] - 2021-08-30
### Other
- Improves documentation.

## [0.3.0] - 2021-08-30
### Changed
- build(deps): bump golang.org/x/net

### Other
- Only pulls image if not already present.
- Removes mock engine container on stop.
- Renames 'mock' subcommand to 'up'.

## [0.2.0] - 2021-08-30

### Other
- Passes through log level to mock engine.
- Sets default log level to info.

## [0.1.0] - 2021-08-30

### Other
- Initial commit.
