#!/usr/bin/env bash

set -xeuo pipefail

helm init --client-only

#####
# set up the repo dir, and package up all charts
#####
CHARTS_REPO=${CHARTS_REPO:-"https://athens.blob.core.windows.net"}
CHARTS_BUCKET=charts
REPO_DIR=bin/charts # This is where we do the charge merge and dirty things up, not the source chart directory
mkdir -p $REPO_DIR
echo "entering $REPO_DIR"
cd $REPO_DIR
if curl --output /dev/null --silent --head --fail ${CHARTS_REPO}/${CHARTS_BUCKET}/index.yaml; then
  echo "downloading existing index.yaml"
  curl -sLO ${CHARTS_REPO}/${CHARTS_BUCKET}/index.yaml
fi

#####
# package the charts
#####
for dir in `ls ../../charts`;do
    if [ ! -f ../../charts/$dir/Chart.yaml ];then
        echo "skipping $dir because it lacks a Chart.yaml file"
    else
        echo "packaging $dir"
        helm dep build ../../charts/$dir
        helm package ../../charts/$dir
    fi
done

if [ -f $REPO_DIR/index.yaml ]; then
  echo "merging with existing index.yaml"
  helm repo index --url "$CHARTS_REPO/$CHARTS_BUCKET" --merge index.yaml .
else
  echo "generating new index.yaml"
  helm repo index .
fi

#####
# upload to Azure blob storage
#####
if [ ! -v AZURE_STORAGE_CONNECTION_STRING ]; then
    echo "AZURE_STORAGE_CONNECTION_STRING env var required to publish"
    exit 1
fi
echo "uploading from $PWD"
az storage blob upload-batch --destination $CHARTS_BUCKET --source .
