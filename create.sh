#!/usr/bin/env bash

cd linux
upx zapsi_demodata_service_linux
cd ..

docker rmi -f petrjahoda/zapsi_demodata_service:latest
docker build -t petrjahoda/zapsi_demodata_service:latest .
docker push petrjahoda/zapsi_demodata_service:latest

docker rmi -f petrjahoda/zapsi_demodata_service:2020.4.1
docker build -t petrjahoda/zapsi_demodata_service:2020.4.1 .
docker push petrjahoda/zapsi_demodata_service:2020.4.1
