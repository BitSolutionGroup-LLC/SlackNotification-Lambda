#!/bin/bash
echo removing old binary....
rm slack-notify slack-notify.zip -f
echo creating binary.... slack-notify
GO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -i -o slack-notify main.go
echo ziping binary....
zip slack-notify.zip slack-notify
echo uploading zip file to s3....
aws s3 cp slack-notify.zip s3://path-to-store
echo successfully uploaded slack-notify.zip
