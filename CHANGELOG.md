# Alarm Service Changelog

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/).

Please note, that this project, while following numbering syntax, it DOES NOT
adhere to [Semantic Versioning](http://semver.org/spec/v2.0.0.html) rules.

## Types of changes

* ```Added``` for new features.
* ```Changed``` for changes in existing functionality.
* ```Deprecated``` for soon-to-be removed features.
* ```Removed``` for now removed features.
* ```Fixed``` for any bug fixes.
* ```Security``` in case of vulnerabilities.

## [2020.4.2.17] - 2020-11-17

### Changed
- added creating additional orders, products, users, downtimes
- updating default downtime, order and user records

## [2020.4.1.26] - 2020-10-26

### Fixed
- fixed leaking goroutine bug when opening sql connections, the right way is this way

## [2020.4.1.1] - 2020-10-1

### Changed
- constant number for creating devices and workplaces

## [2020.3.3.28] - 2020-09-28

### Changed
- updated readme.md
- updated dockerfile
- updated create.sh script

### Added
- creating terminals and link them with workplaces
- creating proper workplace workshifts

## [2020.3.2.22] - 2020-08-29

### Changed
- functions naming changed to idiomatic go (ThisFunction -> thisFunction)

## [2020.3.2.22] - 2020-08-22

### Added
- automatic go get -u all when creating docker image


## [2020.3.2.4] - 2020-08-04

### Changed
- update to latest libraries and latest database changes
- removed all about config and logging to file

## [2020.1.2.29] - 2020-02-29

### Change
- update for latest database changes
- minor changes after testing new postgres, mariadb and mssql
- when searching for active devices, changed from "true" to "1"

## [2020.1.1.1] - 2020-01-01

### Added
- creates 20 devices and 20 workplace if not present
- generates pseudorandom analog and digital data for those 20 devices
