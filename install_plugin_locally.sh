#!/bin/bash

go build
pluginDir=~/.terraform.d/plugins/terraform-example.com/test/cidr-reservator/0.0.1/darwin_arm64/
mkdir -p $pluginDir
cp terraform-provider-cidr-reservator $pluginDir/