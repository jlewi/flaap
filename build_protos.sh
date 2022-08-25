#!/bin/bash
#
# Compile the protocol buffers
set -ex
SRC_DIR=protos

# Delete any protoco buffers already generated
find ./ -name "*pb2.py" -exec rm {} ";"
find ./ -name "*pb2_grpc.py" -exec rm {} ";"
find ./ -name "*pb.go" -exec rm {} ";"
find ./ -name "*pb.gw.go" -exec rm {} ";"

rm -rf go/protos

#********************************************
# Download dependencies
#*******************************************
# This is ideally where a tool like buf could help.

if [ ! -d ".build/grpc" ]; then
    git clone --depth 1 -b v1.48.0 git@github.com:grpc/grpc.git .build/grpc
fi

if [ ! -d ".build/federated" ]; then
    git clone --depth 1 -b v0.33.0 git@github.com:tensorflow/federated.git .build/federated
fi

INCLUDE="-I${SRC_DIR} -I.build/grpc/src/proto -I.build/federated/"

BUILDIR=.build/protos
mkdir -p ${BUILDIR}

SRCS="protos/flaap/taskstore.proto"

# Since the TFF protos don't use option go_package to specify the TFF import
# we need to generate --go_opt=M${PROTO_FILE}=${GO_IMPORT} and
# --go_grpc_opt=M${PROTO_FILE}=${GO_IMPORT} for each file.
# TODO(jeremy): Another case where the buf proto tool would potentially be useful.
TFF_SRCS=""
TFF_NAMES=( "computation.proto" "executor.proto" )
TFF_GO_OPT=""
TFF_GO_GRPC_OPT=""
# The GOLANG import path to use for the protos.
TFF_GOIMPORT="github.com/jlewi/flaap/go/protos/tff/v0"
for f in "${TFF_NAMES[@]}"; do
  if [ ! -z "${TFF_SRCS}" ]; then
    TFF_SRCS+=" "
    TFF_GO_OPT+=" "
    TFF_GO_GRPC_OPT+=" "
  fi
  full_path=".build/federated/tensorflow_federated/proto/v0/${f}"
  TFF_SRCS+=${full_path}
  import_path="tensorflow_federated/proto/v0/${f}"
  TFF_GO_OPT+="--go_opt=M${import_path}=${TFF_GOIMPORT}"
  TFF_GO_GRPC_OPT+="--go-grpc_opt=M${import_path}=${TFF_GOIMPORT}"
done


# Generate the Python protocol buffer files
python3 -m grpc_tools.protoc \
  ${INCLUDE} \
  --python_out=./py \
  --grpc_python_out=./py \
  ${SRCS}

# HACK generate protos for computation.prot
# protoc ${INCLUDE} \
#     --go_out ${BUILDIR} --go_opt paths=import \
#     --go_opt=Mtensorflow_federated/proto/v0/computation.proto=${TFF_GOIMPORT} \
#     --go-grpc_out ${BUILDIR} --go-grpc_opt paths=import \
#     --go-grpc_opt=Mtensorflow_federated/proto/v0/computation.proto=${TFF_GOIMPORT} \
#     tensorflow_federated/proto/v0/computation.proto

# echo exiting early do not commit
# exit 0

# Generate golang protocol buffers
# For GoLang we need to generate protocol buffers for TFF as well as taskstore
# because the TFF repository doesn't publish them.
protoc ${INCLUDE} \
    --go_out ${BUILDIR} --go_opt paths=import \
    ${TFF_GO_OPT} \
    --go-grpc_out ${BUILDIR} --go-grpc_opt paths=import \
    ${TFF_GO_GRPC_OPT} \
    ${SRCS} ${TFF_SRCS}

# Hack move the generated go files to the proper location
mv -f .build/protos/github.com/jlewi/flaap/go/protos/ \
    go/

