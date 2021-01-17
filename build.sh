#!/bin/sh

PROG=$0
CURRENT_DIR=$PWD

ECR_REGION=`cat docker-manifest.json | jq -j .ecr_region`
CONTAINER_VERSION=`cat docker-manifest.json | jq -j .container_version`
CONTAINER_NAME=`cat docker-manifest.json | jq -j .container_name`

function printSyntax(){
   echo "ERROR: invalid parameters."
   echo "Syntaxe :"
   echo "\t$PROG [local|ecr] <optional extra build parameters>"
   echo
   echo " - local : build the container on your laptop and don't push on AWS ECR"
   echo " - ecr   : build the container on your laptop and push on AWS ECR"
}

function buildECRRepoURI(){
   # Launch the command to verify if the registry exist. If not send error code and create it
   aws ecr describe-repositories --region ${ECR_REGION} --repository-name ${CONTAINER_NAME}  --output json --query 'repositories[0].repositoryUri'
   if [ $? -ne 0 ]
   then
      # Create repository on ECR :
      aws ecr create-repository --region ${ECR_REGION} --repository-name ${CONTAINER_NAME} # obsolete--image-tag-mutability MUTABLE --image-scanning-configuration scanOnPush=true
   fi 
   REPO_URI=`aws ecr describe-repositories --region ${ECR_REGION} --repository-name ${CONTAINER_NAME}  --output json --query 'repositories[0].repositoryUri' | sed 's:"::g'`

   return buildECRRepoURI
}

function buildContainer(){
   echo "Try to build : ${CONTAINER_NAME} store in ${ECR_REGION}"
   ECR_CONTAINER_TAG=${REPO_URI}:${CONTAINER_VERSION}
   ECR_CONTAINER_TAG_LATEST=${REPO_URI}:latest
   docker build ./ -t ${ECR_CONTAINER_TAG} -t ${ECR_CONTAINER_TAG_LATEST} ${BUILD_PARAMETERS}
   DOCKER_ERROR_CODE=$?
   if [ ${DOCKER_ERROR_CODE} -ne 0 ]; then
      echo "Create Docker image Error : (${DOCKER_ERROR_CODE})..."
      exit ${DOCKER_ERROR_CODE}
   fi
}

function pushContainerToECR(){
   echo "Upload Image ..."
   eval $(aws ecr get-login --no-include-email --region ${ECR_REGION})
   echo "docker push ${ECR_CONTAINER_TAG} ..."
   docker push ${ECR_CONTAINER_TAG} 
   docker push "${ECR_CONTAINER_TAG_LATEST} ..."
   docker push ${ECR_CONTAINER_TAG_LATEST}

   echo "List ECR Images ..."
   aws ecr list-images --region ${ECR_REGION}  --repository-name ${CONTAINER_NAME}

   echo "Pull latest command ..."
   echo "docker pull ${ECR_CONTAINER_TAG_LATEST}"
}

if [ $# -lt 1 ]
then
   printSyntax
   exit 1
fi

[ $# -eq 2 ] && BUILD_PARAMETERS=$2

case $1 in
   "local")
      echo "Build local"
      REPO_URI=${CONTAINER_NAME}
      buildContainer
      ;;
   "ecr")
      echo "Build and push to ecr"
      REPO_URI=buildECRRepoURI
      buildContainer
      pushContainerToECR
      ;;
   *)
      # TODO: Add new registry
      printSyntax
      exit 1
      ;;
esac

echo "THE END"