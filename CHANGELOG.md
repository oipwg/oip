# Changelog
## [Unreleased] - Future Release
### Changed
- Refactor Edit error handling & unsquash function to support `remove` operations

## [mlg-1.3.1] - Jun-26-2019
### Added
- `/oip042/edit/search` API endpoint

### Changed
- Increase Index Shard count for `oip042_artifacts` to `3` (2 -> 3)
- Increase Index Shard count for `multiparts` to `3` (2 -> 3)
- Increase Index Shard count for `blocks` to `5` (2 -> 5)
- Increase Index Shard count for `transactions` to `6` (2 -> 6)
- Decreased all other indexes to only have a single shard (2 -> 1)
- Disabled Shard Replicas
- Modified Multipart processing to only occur after `initialSync` is complete
- Changed Default RAM for ElasticSearch to 1/4th of the total system RAM
- Changed the Bulk Indexer to only write 10MB of data at a time instead of 80MB

## [mlg-1.3.0] - Jun-13-2019
### Added
- Support for Block Reorganizations

### Fixed
- Occasionally the EventBus locks up causing all new Blocks and Records to cease being processed. 

## [mlg-1.2.2] - May-30-2019
### Fixed
- Use `mainnet` instead of `livenet` as default network name

### Changed
- Docker Healthcheck command uses `/sync/status` to check health

## [mlg-1.2.1] - May-29-2019
### Fixed
- Fixed memory leak in Multipart logic

## [mlg-1.2.0] - May-21-2019
### Added
- OIP 042 Edit Functionality
- Edit API Endpoints

### Changed
- Filter Blacklisted Records from API Results

## [mlg-1.1.5] - Apr-29-2019
### Added
- Add SpatialData index field

## [mlg-1.1.4] - Apr-17-2019
### Added
- Build/Publish Docker Containers merged into main Repo

### Changed
- Increase ElasticSearch default Heap to 2GB

## [mlg-1.1.3] - Apr-14-2019
### Changed
- Uses internal Dockerfile to build Binaries

## [mlg-1.1.2] - Apr-05-2019
### Changed
- Merged in official Multipart change from GitHub

### Fixed
- Signature Validation crash on Mainnet

## [mlg-1.1.1] - Feb-28-2019
### Changed
- Loop Multiparts to allow >10k pending multiparts

### Removed
- Disabled markStale() Multiparts method

## [mlg-1.1.0] - Feb-06-2019
### Added
- Added CORS to httpapi
- Added new Fields to Elasticsearch mappings to support Property Records
- Added Remote Blacklist Support
- Added gzip compression for API responses
- Added API pagination

### Changed
- Updated API mapping

## [mlg-1.0.1] - Feb-04-2019
### Fixed
- Allow Files to contain a location at `record.storage.files[i].location`

## [mlg-1.0.0] - Jan-08-2019
### Added
- Initial Release of Docker Container