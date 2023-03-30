#!/bin/bash


# check for sopm files
# if the input is not defined use any sopm in the current directory
[ "${INPUT_SOPM}" == "" ] && INPUT_SOPM=$(ls *.sopm)
if [ "${INPUT_SOPM}" == "" ]; then
  echo "no sopm file found"
  exit 2
fi

# check if version and name have been given
if [ "${INPUT_VERSION}" == "" ]; then
  echo "input version missing"
  exit 2
fi
if [ "${INPUT_NAME}" == "" ]; then
  echo "input name missing"
  exit 2
fi

# build the opm file
/opmbuilder build --version ${INPUT_VERSION} --output ${INPUT_NAME}-${INPUT_VERSION}.opm ${INPUT_SOPM}
