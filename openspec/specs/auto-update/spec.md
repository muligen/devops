## ADDED Requirements

### Requirement: Agent checks for updates

The Agent (main process) SHALL periodically check for updates.

#### Scenario: Scheduled update check
- **WHEN** `check_interval` elapses (default 1 hour)
- **THEN** Agent queries Server for latest version
- **AND** Agent checks if Worker processes are idle

#### Scenario: Skip update during tasks
- **WHEN** update check is due but Worker processes are busy
- **THEN** Agent skips update check
- **AND** Agent retries on next interval

### Requirement: Agent downloads and verifies updates

The Agent SHALL download and verify update packages.

#### Scenario: Update available
- **WHEN** Server reports newer version available
- **THEN** Agent downloads update package from provided URL
- **AND** Agent verifies file hash (SHA256)
- **AND** Agent verifies file signature

#### Scenario: Verification failure
- **WHEN** hash or signature verification fails
- **THEN** Agent discards downloaded file
- **AND** Agent logs verification error
- **AND** Agent reports failure to Server

### Requirement: Agent performs hot update

The Agent SHALL update Worker processes without full restart.

#### Scenario: Update Worker processes
- **WHEN** update is verified and Workers are idle
- **THEN** Agent stops Worker processes
- **AND** Agent replaces Worker executables
- **AND** Agent restarts Worker processes
- **AND** Agent reports update success to Server

#### Scenario: Update failure rollback
- **WHEN** Worker fails to start after update
- **THEN** Agent restores previous version
- **AND** Agent restarts Workers with old version
- **AND** Agent reports update failure to Server

### Requirement: Server manages versions

The Server SHALL manage Agent versions.

#### Scenario: Version registration
- **WHEN** admin uploads new version
- **THEN** Server stores version metadata
- **AND** Server stores update package in object storage
- **AND** Server calculates and stores file hash

#### Scenario: Version query
- **WHEN** Agent queries for updates
- **THEN** Server returns latest version for each component
- **AND** Server returns download URL and file hash

### Requirement: Server distributes update packages

The Server SHALL distribute update packages via signed URLs.

#### Scenario: Signed URL generation
- **WHEN** Agent requests update download
- **THEN** Server generates time-limited signed URL
- **AND** URL expires after 1 hour
- **AND** URL is served from object storage (MinIO)

### Requirement: Agent reports update status

The Agent SHALL report update status to Server.

#### Scenario: Update success report
- **WHEN** update completes successfully
- **THEN** Agent sends `update.success` message
- **AND** message contains new version number
- **AND** Server updates Agent's current version

#### Scenario: Update failure report
- **WHEN** update fails
- **THEN** Agent sends `update.failed` message
- **AND** message contains error details
- **AND** Server logs the failure

### Requirement: Main process remains stable

The Agent main process SHALL rarely require updates.

#### Scenario: Main process update
- **WHEN** main process update is required
- **THEN** Agent downloads new main process
- **AND** Agent schedules restart via Windows Service Manager
- **AND** Agent minimizes downtime (target: < 5 seconds)
