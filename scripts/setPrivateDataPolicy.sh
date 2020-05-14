#!/bin/bash

PATH_TO_COLLECTION_FILE=${PWD}/chaincode/privateCollections.json
export PRIVATE_COLLECTION_POLICY="--collections-config ${PATH_TO_COLLECTION_FILE}"