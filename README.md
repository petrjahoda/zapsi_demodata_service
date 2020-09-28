[![developed_using](https://img.shields.io/badge/developed%20using-Jetbrains%20Goland-lightgrey)](https://www.jetbrains.com/go/)
<br/>
![GitHub](https://img.shields.io/github/license/petrjahoda/zapsi_demodata_service)
[![GitHub last commit](https://img.shields.io/github/last-commit/petrjahoda/zapsi_demodata_service)](https://github.com/petrjahoda/zapsi_demodata_service/commits/master)
[![GitHub issues](https://img.shields.io/github/issues/petrjahoda/zapsi_demodata_service)](https://github.com/petrjahoda/zapsi_demodata_service/issues)
<br/>
![GitHub language count](https://img.shields.io/github/languages/count/petrjahoda/zapsi_demodata_service)
![GitHub top language](https://img.shields.io/github/languages/top/petrjahoda/zapsi_demodata_service)
![GitHub repo size](https://img.shields.io/github/repo-size/petrjahoda/zapsi_demodata_service)
<br/>
[![Docker Pulls](https://img.shields.io/docker/pulls/petrjahoda/zapsi_demodata_service)](https://hub.docker.com/r/petrjahoda/zapsi_demodata_service)
[![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/petrjahoda/zapsi_demodata_service?sort=date)](https://hub.docker.com/r/petrjahoda/zapsi_demodata_service/tags)
<br/>
[![developed_using](https://img.shields.io/badge/database-PostgreSQL-red)](https://www.postgresql.org) [![developed_using](https://img.shields.io/badge/runtime-Docker-red)](https://www.docker.com)

# Zapsi Demodata Service
## Description
Go service that generates pseudorandom data like "from Zapsi" devices and insert them to database.

## Installation Information
Install under docker runtime using [this dockerfile image](https://github.com/petrjahoda/system/tree/master/latest) with this command: ```docker-compose up -d```

## Implementation Information
Check the software running with this command: ```docker stats```. <br/>
Zapsi_demodata_service has to be running.


## Additional information
* creates 20 devices with device ports
* creates 20 workplaces with workplace ports
* creates 20 terminals and link them with workplaces
* links workplaces with workshifts


© 2020 Petr Jahoda
