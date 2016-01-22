#!/bin/bash

# Copyright 2015 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# kubernetes-e2e-{gce, gke, gke-ci} jobs: This script is triggered by
# the kubernetes-build job, or runs every half hour. We abort this job
# if it takes more than 75m. As of initial commit, it typically runs
# in about half an hour.
#
# The "Workspace Cleanup Plugin" is installed and in use for this job,
# so the ${WORKSPACE} directory (the current directory) is currently
# empty.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Join all args with |
#   Example: join_regex_allow_empty a b "c d" e  =>  a|b|c d|e
function join_regex_allow_empty() {
    local IFS="|"
    echo "$*"
}

# Join all args with |, butin case of empty result prints "EMPTY\sSET" instead.
#   Example: join_regex_no_empty a b "c d" e  =>  a|b|c d|e
#            join_regex_no_empty => EMPTY\sSET
function join_regex_no_empty() {
    local IFS="|"
    if [ -z "$*" ]; then
        echo "EMPTY\sSET"
    else
        echo "$*"
    fi
}

# Properly configure globals for an upgrade step in a GKE or GCE upgrade suite
#
# These suites:
#   step1: launch a cluster at $old_version,
#   step2: upgrades the master to $new_version,
#   step3: runs $old_version e2es,
#   step4: upgrades the rest of the cluster,
#   step5: runs $old_version e2es again, then
#   step6: runs $new_version e2es and tears down the cluster.
#
# Assumes globals:
#   $JOB_NAME
#   $KUBERNETES_PROVIDER
#   $GKE_DEFAULT_SKIP_TESTS
#   $GCE_DEFAULT_SKIP_TESTS
#   $GCE_FLAKY_TESTS
#   $GCE_SLOW_TESTS
#   $GKE_FLAKY_TESTS
#
# Args:
#   $1 old_version:  the version to deploy a cluster at, and old e2e tests to run
#                    against the upgraded cluster (should be something like
#                    'release/latest', to work with JENKINS_PUBLISHED_VERSION logic)
#   $2 new_version:  the version to upgrade the cluster to, and new e2e tests to run
#                    against the upgraded cluster (should be something like
#                    'ci/latest', to work with JENKINS_PUBLISHED_VERSION logic)
#   $3 cluster_name: determines E2E_CLUSTER_NAME and E2E_NETWORK
#   $4 project:      determines PROJECT

function configure_upgrade_step() {
  local -r old_version="$1"
  local -r new_version="$2"
  local -r cluster_name="$3"
  local -r project="$4"

  [[ "${JOB_NAME}" =~ .*-(step[1-6])-.* ]] || {
    echo "JOB_NAME ${JOB_NAME} is not a valid upgrade job name, could not parse"
    exit 1
  }
  local -r step="${BASH_REMATCH[1]}"

  local -r gce_test_args="--ginkgo.skip=$(join_regex_allow_empty \
        ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
        ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
        ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
        )"
  local -r gke_test_args="--ginkgo.skip=$(join_regex_allow_empty \
        ${GKE_DEFAULT_SKIP_TESTS[@]:+${GKE_DEFAULT_SKIP_TESTS[@]}} \
        ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
        ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
        ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
        ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
        )"

  if [[ "${KUBERNETES_PROVIDER}" == "gke" ]]; then
    DOGFOOD_GCLOUD="true"
    GKE_API_ENDPOINT="https://test-container.sandbox.googleapis.com/"
  fi

  E2E_CLUSTER_NAME="$cluster_name"
  E2E_NETWORK="$cluster_name"
  PROJECT="$project"

  case $step in
    step1)
      # Deploy at old version
      JENKINS_PUBLISHED_VERSION="${old_version}"

      E2E_UP="true"
      E2E_TEST="false"
      E2E_DOWN="false"

      if [[ "${KUBERNETES_PROVIDER}" == "gke" ]]; then
        E2E_SET_CLUSTER_API_VERSION=y
      fi
      ;;

    step2)
      # Use upgrade logic of version we're upgrading to.
      JENKINS_PUBLISHED_VERSION="${new_version}"
      JENKINS_FORCE_GET_TARS=y

      E2E_OPT="--check_version_skew=false"
      E2E_UP="false"
      E2E_TEST="true"
      E2E_DOWN="false"
      GINKGO_TEST_ARGS="--ginkgo.focus=Cluster\sUpgrade.*upgrade-master --upgrade-target=${new_version}"
      ;;

    step3)
      # Run old e2es
      JENKINS_PUBLISHED_VERSION="${old_version}"
      JENKINS_FORCE_GET_TARS=y

      E2E_OPT="--check_version_skew=false"
      E2E_UP="false"
      E2E_TEST="true"
      E2E_DOWN="false"

      if [[ "${KUBERNETES_PROVIDER}" == "gke" ]]; then
        GINKGO_TEST_ARGS="${gke_test_args}"
      else
        GINKGO_TEST_ARGS="${gce_test_args}"
      fi
      ;;

    step4)
      # Use upgrade logic of version we're upgrading to.
      JENKINS_PUBLISHED_VERSION="${new_version}"
      JENKINS_FORCE_GET_TARS=y

      E2E_OPT="--check_version_skew=false"
      E2E_UP="false"
      E2E_TEST="true"
      E2E_DOWN="false"
      GINKGO_TEST_ARGS="--ginkgo.focus=Cluster\sUpgrade.*upgrade-cluster --upgrade-target=${new_version}"
      ;;

    step5)
      # Run old e2es
      JENKINS_PUBLISHED_VERSION="${old_version}"
      JENKINS_FORCE_GET_TARS=y

      E2E_OPT="--check_version_skew=false"
      E2E_UP="false"
      E2E_TEST="true"
      E2E_DOWN="false"

      if [[ "${KUBERNETES_PROVIDER}" == "gke" ]]; then
        GINKGO_TEST_ARGS="${gke_test_args}"
      else
        GINKGO_TEST_ARGS="${gce_test_args}"
      fi
      ;;

    step6)
      # Run new e2es
      JENKINS_PUBLISHED_VERSION="${new_version}"
      JENKINS_FORCE_GET_TARS=y

      # TODO(15011): these really shouldn't be (very) version skewed, but
      # because we have to get ci/latest again, it could get slightly out of
      # whack.
      E2E_OPT="--check_version_skew=false"
      E2E_UP="false"
      E2E_TEST="true"
      E2E_DOWN="true"

      if [[ "${KUBERNETES_PROVIDER}" == "gke" ]]; then
        GINKGO_TEST_ARGS="${gke_test_args}"
      else
        GINKGO_TEST_ARGS="${gce_test_args}"
      fi
      ;;
  esac
}

echo "--------------------------------------------------------------------------------"
echo "Initial Environment:"
printenv | sort
echo "--------------------------------------------------------------------------------"

if [[ "${CIRCLECI:-}" == "true" ]]; then
    JOB_NAME="circleci-${CIRCLE_PROJECT_USERNAME}-${CIRCLE_PROJECT_REPONAME}"
    BUILD_NUMBER=${CIRCLE_BUILD_NUM}
    WORKSPACE=`pwd`
else
    # Jenkins?
    export HOME=${WORKSPACE} # Nothing should want Jenkins $HOME
fi

# Additional parameters that are passed to hack/e2e.go
E2E_OPT=${E2E_OPT:-""}

# Set environment variables shared for all of the GCE Jenkins projects.
if [[ ${JOB_NAME} =~ ^kubernetes-.*-gce ]]; then
  KUBERNETES_PROVIDER="gce"
  : ${E2E_MIN_STARTUP_PODS:="1"}
  : ${E2E_ZONE:="us-central1-f"}
  : ${NUM_NODES_PARALLEL:="6"}  # Number of nodes required to run all of the tests in parallel

elif [[ ${JOB_NAME} =~ ^kubernetes-.*-gke ]]; then
  KUBERNETES_PROVIDER="gke"
  : ${E2E_ZONE:="us-central1-f"}
elif [[ ${JOB_NAME} =~ ^kubernetes-.*-aws ]]; then
  KUBERNETES_PROVIDER="aws"
  : ${E2E_MIN_STARTUP_PODS:="1"}
  : ${E2E_ZONE:="us-east-1a"}
  : ${NUM_NODES_PARALLEL:="6"}  # Number of nodes required to run all of the tests in parallel
fi

if [[ "${KUBERNETES_PROVIDER}" == "aws" ]]; then
  if [[ "${PERFORMANCE:-}" == "true" ]]; then
    : ${MASTER_SIZE:="m3.xlarge"}
    : ${NUM_NODES:="100"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=\[Performance\]"}
  else
    : ${MASTER_SIZE:="m3.large"}
    : ${NODE_SIZE:="m3.large"}
    : ${NUM_NODES:="3"}
  fi
fi

# CURRENT_RELEASE_PUBLISHED_VERSION is the JENKINS_PUBLISHED_VERSION for the
# release we are currently pointing our release testing infrastructure at.
# When 1.2.0-beta.0 comes out, e.g., this will become "ci/latest-1.2"
CURRENT_RELEASE_PUBLISHED_VERSION="ci/latest-1.1"

# Specialized to skip when running reboot tests.
REBOOT_SKIP_TESTS=(
    "Restart\sshould\srestart\sall\snodes"
    "\[Example\]"
    )

# Specialized tests which should be skipped by default for projects.
GCE_DEFAULT_SKIP_TESTS=(
    "${REBOOT_SKIP_TESTS[@]}"
    "\[Skipped\]"
    "Reboot"
    "ServiceLoadBalancer"
    )

# Tests which cannot be run on GKE, e.g. because they require
# master ssh access.
GKE_REQUIRED_SKIP_TESTS=(
    "Nodes"
    "Etcd\sFailure"
    "MasterCerts"
    "experimental\sresource\susage\stracking" # Expect --max-pods=100
    "ServiceLoadBalancer" # issue: #16602
    "Shell"
    # Alpha features, remove from skip when these move to beta
    "Daemon\sset"
    "Deployment"
    )

# Tests wchich are known to be flaky on GKE
GKE_FLAKY_TESTS=(
    "NodeOutOfDisk"
  )

# Specialized tests which should be skipped by default for GKE.
GKE_DEFAULT_SKIP_TESTS=(
    "Autoscaling\sSuite"
    # Perf test, slow by design
    "resource\susage\stracking"
    "${GKE_REQUIRED_SKIP_TESTS[@]}"
    )

# Tests which cannot be run on AWS.
AWS_REQUIRED_SKIP_TESTS=(
    "experimental\sresource\susage\stracking" # Expect --max-pods=100
    "GCE\sL7\sLoadBalancer\sController" # GCE L7 loadbalancing
)


# Tests which kills or restarts components and/or nodes.
DISRUPTIVE_TESTS=(
    "DaemonRestart"
    "Etcd\sfailure"
    "Nodes\sNetwork"
    "Nodes\sResize"
    "Reboot"
    "Services.*restarting"
)

# The following tests are known to be flaky, and are thus run only in their own
# -flaky- build variants.
GCE_FLAKY_TESTS=(
    "GCE\sL7\sLoadBalancer\sController" # issue: #17518
    "DaemonRestart\sController\sManager" # issue: #17829
    "Resource\susage\sof\ssystem\scontainers" # issue: #13931
    "NodeOutOfDisk" # issue: #17687
    "Cluster\slevel\slogging\susing\sElasticsearch" # issue: #17873
    )

# The following tests are known to be slow running (> 2 min), and are
# thus run only in their own -slow- build variants.  Note that tests
# can be slow by explicit design (e.g. some soak tests), or slow
# through poor implementation.  Please indicate which applies in the
# comments below, and for poorly implemented tests, please quote the
# issue number tracking speed improvements.
GCE_SLOW_TESTS=(
    # Before enabling this loadbalancer test in any other test list you must
    # make sure the associated project has enough quota. At the time of this
    # writing a GCE project is allowed 3 backend services by default. This
    # test requires at least 5.
    "GCE\sL7\sLoadBalancer\sController"               # 10 min,       file: ingress.go,              slow by design
    "SchedulerPredicates\svalidates\sMaxPods\slimit " # 8 min,        file: scheduler_predicates.go, PR:    #13315
    "Nodes\sResize"                                   # 3 min 30 sec, file: resize_nodes.go,         issue: #13323
    "resource\susage\stracking"                       # 1 hour,       file: kubelet_perf.go,         slow by design
    "monotonically\sincreasing\srestart\scount"       # 1.5 to 5 min, file: pods.go,                 slow by design
    "Garbage\scollector\sshould"                      # 7 min,        file: garbage_collector.go,    slow by design
    "KubeProxy\sshould\stest\skube-proxy"             # 9 min 30 sec, file: kubeproxy.go,            issue: #14204
    "cap\sback-off\sat\sMaxContainerBackOff"          # 20 mins       file: manager.go,              PR:    #12648
    )

# Tests which are not able to be run in parallel.
GCE_PARALLEL_SKIP_TESTS=(
    "GCE\sL7\sLoadBalancer\sController" # namespaced watch flakes, issue: #17805
    "Nodes\sNetwork"
    "MaxPods"
    "Resource\susage\sof\ssystem\scontainers"
    "SchedulerPredicates"
    "resource\susage\stracking"
    "NodeOutOfDisk"
    "${DISRUPTIVE_TESTS[@]}"
    )

# Tests which are known to be flaky when run in parallel.
GCE_PARALLEL_FLAKY_TESTS=(
    "DaemonRestart"
    "Elasticsearch"
    "Namespaces.*should\sdelete\sfast"
    "ServiceAccounts"
    "Services.*identically\snamed" # error waiting for reachability, issue: #16285
    )

# Tests that should not run on soak cluster.
GCE_SOAK_CONTINUOUS_SKIP_TESTS=(
    "GCE\sL7\sLoadBalancer\sController" # issue: #17119
    "Density.*30\spods"
    "Elasticsearch"
    "external\sload\sbalancer"
    "identically\snamed\sservices"
    "network\spartition"
    "Services.*Type\sgoes\sfrom"
    "${DISRUPTIVE_TESTS[@]}"       # avoid component restarts.
    )

GCE_RELEASE_SKIP_TESTS=(
    )

# Define environment variables based on the Jenkins project name.
# NOTE: Not all jobs are defined here. The hack/jenkins/e2e.sh in master and
# release branches defines relevant jobs for that particular version of
# Kubernetes.
case ${JOB_NAME} in
  # Runs all non-flaky, non-slow tests on GCE, sequentially.
  kubernetes-e2e-gce)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e"}
    : ${E2E_NETWORK:="e2e-gce"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX="e2e-gce"}
    : ${PROJECT:="k8s-jkns-e2e-gce"}
    : ${ENABLE_DEPLOYMENTS:=true}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    ;;

  # Runs all non-flaky, non-slow tests on AWS, sequentially.
  kubernetes-e2e-aws)
    : ${E2E_CLUSTER_NAME:="jenkins-aws-e2e"}
    : ${E2E_DOWN:="false"}
    : ${E2E_NETWORK:="e2e-aws"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          ${AWS_REQUIRED_SKIP_TESTS[@]:+${AWS_REQUIRED_SKIP_TESTS[@]}} \
	  )"}
    : ${KUBE_GCE_INSTANCE_PREFIX="e2e-aws"}
    : ${PROJECT:="k8s-jkns-e2e-aws"}
    : ${ENABLE_DEPLOYMENTS:=true}
    ;;

  # Runs only the examples tests on GCE.
  kubernetes-e2e-gce-examples)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e-examples"}
    : ${E2E_NETWORK:="e2e-examples"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=\[Example\]"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-examples"}
    : ${PROJECT:="kubernetes-jenkins"}
    ;;

  # Runs only the autoscaling tests on GCE.
  kubernetes-e2e-gce-autoscaling)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e-autoscaling"}
    : ${E2E_NETWORK:="e2e-autoscaling"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=\[Autoscaling\]"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-autoscaling"}
    : ${PROJECT:="k8s-jnks-e2e-gce-autoscaling"}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    : ${ENABLE_DEPLOYMENTS:=true}
    # Override GCE default for cluster size autoscaling purposes.
    ENABLE_CLUSTER_MONITORING="googleinfluxdb"
    ADMISSION_CONTROL="NamespaceLifecycle,InitialResources,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota"
    ;;

  # Runs the flaky tests on GCE, sequentially.
  kubernetes-e2e-gce-flaky)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e-flaky"}
    : ${E2E_NETWORK:="e2e-flaky"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ) --ginkgo.focus=$(join_regex_no_empty \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-flaky"}
    : ${PROJECT:="k8s-jkns-e2e-gce-flaky"}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    ;;

  # Runs slow tests on GCE, sequentially.
  kubernetes-e2e-gce-slow)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e-slow"}
    : ${E2E_NETWORK:="e2e-slow"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=$(join_regex_no_empty \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-slow"}
    : ${PROJECT:="k8s-jkns-e2e-gce-slow"}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    ;;

  # Runs a subset of tests on GCE in parallel. Run against all pending PRs.
  kubernetes-pull-build-test-e2e-gce)
    : ${E2E_CLUSTER_NAME:="jnks-e2e-gce-${NODE_NAME}-${EXECUTOR_NUMBER}"}
    : ${E2E_NETWORK:="e2e-gce-${NODE_NAME}-${EXECUTOR_NUMBER}"}
    : ${GINKGO_PARALLEL:="y"}
    # This list should match the list in kubernetes-e2e-gce-parallel.
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_PARALLEL_SKIP_TESTS[@]:+${GCE_PARALLEL_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_PARALLEL_FLAKY_TESTS[@]:+${GCE_PARALLEL_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-gce-${NODE_NAME}-${EXECUTOR_NUMBER}"}
    : ${KUBE_GCS_STAGING_PATH_SUFFIX:="-${NODE_NAME}-${EXECUTOR_NUMBER}"}
    : ${PROJECT:="kubernetes-jenkins-pull"}
    : ${ENABLE_DEPLOYMENTS:=true}
    # Override GCE defaults
    NUM_NODES=${NUM_NODES_PARALLEL}
    ;;

  # Runs all non-flaky tests on GCE in parallel.
  kubernetes-e2e-gce-parallel)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e-parallel"}
    : ${E2E_NETWORK:="e2e-parallel"}
    : ${GINKGO_PARALLEL:="y"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_PARALLEL_SKIP_TESTS[@]:+${GCE_PARALLEL_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_PARALLEL_FLAKY_TESTS[@]:+${GCE_PARALLEL_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-test-parallel"}
    : ${PROJECT:="kubernetes-jenkins"}
    : ${ENABLE_DEPLOYMENTS:=true}
    # Override GCE defaults
    NUM_NODES=${NUM_NODES_PARALLEL}
    ;;

  # Runs all non-flaky tests on AWS in parallel.
  kubernetes-e2e-aws-parallel)
    : ${E2E_CLUSTER_NAME:="jenkins-aws-e2e-parallel"}
    : ${E2E_NETWORK:="e2e-parallel"}
    : ${GINKGO_PARALLEL:="y"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_PARALLEL_SKIP_TESTS[@]:+${GCE_PARALLEL_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_PARALLEL_FLAKY_TESTS[@]:+${GCE_PARALLEL_FLAKY_TESTS[@]}} \
          ${AWS_REQUIRED_SKIP_TESTS[@]:+${AWS_REQUIRED_SKIP_TESTS[@]}} \
          )"}
    : ${ENABLE_DEPLOYMENTS:=true}
    # Override AWS defaults.
    NUM_NODES=${NUM_NODES_PARALLEL}
    ;;

  # Runs the flaky tests on GCE in parallel.
  kubernetes-e2e-gce-parallel-flaky)
    : ${E2E_CLUSTER_NAME:="parallel-flaky"}
    : ${E2E_NETWORK:="e2e-parallel-flaky"}
    : ${GINKGO_PARALLEL:="y"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_PARALLEL_SKIP_TESTS[@]:+${GCE_PARALLEL_SKIP_TESTS[@]}} \
          ) --ginkgo.focus=$(join_regex_no_empty \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_PARALLEL_FLAKY_TESTS[@]:+${GCE_PARALLEL_FLAKY_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="parallel-flaky"}
    : ${PROJECT:="k8s-jkns-e2e-gce-prl-flaky"}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    # Override GCE defaults.
    NUM_NODES=${NUM_NODES_PARALLEL}
    ;;

  # Runs only the reboot tests on GCE.
  kubernetes-e2e-gce-reboot)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e-reboot"}
    : ${E2E_NETWORK:="e2e-reboot"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=Reboot"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-reboot"}
    : ${PROJECT:="kubernetes-jenkins"}
    ;;

  # Runs the performance/scalability tests on GCE. A larger cluster is used.
  kubernetes-e2e-gce-scalability)
    : ${E2E_CLUSTER_NAME:="jenkins-gce-e2e-scalability"}
    : ${E2E_NETWORK:="e2e-scalability"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=\[Performance\] \
        --gather-resource-usage=true"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-scalability"}
    : ${PROJECT:="kubernetes-jenkins"}
    # Override GCE defaults.
    MASTER_SIZE="n1-standard-4"
    NODE_SIZE="n1-standard-2"
    NODE_DISK_SIZE="50GB"
    NUM_NODES="100"
    # Reduce logs verbosity
    TEST_CLUSTER_LOG_LEVEL="--v=2"
    # Increase resync period to simulate production
    TEST_CLUSTER_RESYNC_PERIOD="--min-resync-period=12h"
    ;;

  # Sets up the GCE soak cluster weekly using the latest CI release.
  kubernetes-soak-weekly-deploy-gce)
    : ${E2E_CLUSTER_NAME:="gce-soak-weekly"}
    : ${E2E_DOWN:="false"}
    : ${E2E_NETWORK:="gce-soak-weekly"}
    : ${E2E_TEST:="false"}
    : ${E2E_UP:="true"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="gce-soak-weekly"}
    : ${PROJECT:="kubernetes-jenkins"}
    ;;

  # Runs tests on GCE soak cluster.
  kubernetes-soak-continuous-e2e-gce)
    : ${E2E_CLUSTER_NAME:="gce-soak-weekly"}
    : ${E2E_DOWN:="false"}
    : ${E2E_NETWORK:="gce-soak-weekly"}
    : ${E2E_UP:="false"}
    # Clear out any orphaned namespaces in case previous run was interrupted.
    : ${E2E_CLEAN_START:="true"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_SOAK_CONTINUOUS_SKIP_TESTS[@]:+${GCE_SOAK_CONTINUOUS_SKIP_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="gce-soak-weekly"}
    : ${PROJECT:="kubernetes-jenkins"}
    ;;

  kubernetes-e2e-gke-ci)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${CLOUDSDK_BUCKET:="gs://cloud-sdk-build/testing/staging"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="jkns-gke-e2e-ci"}
    : ${E2E_NETWORK:="e2e-gke-ci"}
    : ${E2E_SET_CLUSTER_API_VERSION:=y}
    : ${PROJECT:="k8s-jkns-e2e-gke-ci"}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GKE_DEFAULT_SKIP_TESTS[@]:+${GKE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
          )"}
    ;;

  kubernetes-e2e-gke-ci-reboot)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${CLOUDSDK_BUCKET:="gs://cloud-sdk-build/testing/staging"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="jkns-gke-e2e-ci-reboot"}
    : ${E2E_NETWORK:="e2e-gke-ci-reboot"}
    : ${E2E_SET_CLUSTER_API_VERSION:=y}
    : ${PROJECT:="k8s-jkns-e2e-gke-ci-reboot"}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          "\[Skipped\]" \
          ${GKE_DEFAULT_SKIP_TESTS[@]:+${GKE_DEFAULT_SKIP_TESTS[@]}} \
          ${REBOOT_SKIP_TESTS[@]:+${REBOOT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    ;;

  kubernetes-e2e-gke-flaky)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${CLOUDSDK_BUCKET:="gs://cloud-sdk-build/testing/staging"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="kubernetes-gke-e2e-flaky"}
    : ${E2E_NETWORK:="gke-e2e-flaky"}
    : ${E2E_SET_CLUSTER_API_VERSION:=y}
    : ${PROJECT:="k8s-jkns-e2e-gke-ci-flaky"}
    : ${FAIL_ON_GCP_RESOURCE_LEAK:="true"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=$(join_regex_no_empty \
          ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
          )"}
    ;;

  # Sets up the GKE soak cluster weekly using the latest CI release.
  kubernetes-soak-weekly-deploy-gke)
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="jenkins-gke-soak-weekly"}
    : ${E2E_DOWN:="false"}
    : ${E2E_NETWORK:="gke-soak-weekly"}
    : ${E2E_SET_CLUSTER_API_VERSION:=y}
    : ${JENKINS_PUBLISHED_VERSION:="ci/latest"}
    : ${E2E_TEST:="false"}
    : ${E2E_UP:="true"}
    : ${PROJECT:="kubernetes-jenkins"}
    # Need at least n1-standard-2 nodes to run kubelet_perf tests
    NODE_SIZE="n1-standard-2"
    ;;

  # Runs tests on GKE soak cluster.
  kubernetes-soak-continuous-e2e-gke)
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="jenkins-gke-soak-weekly"}
    : ${E2E_NETWORK:="gke-soak-weekly"}
    : ${E2E_DOWN:="false"}
    : ${E2E_UP:="false"}
    # Clear out any orphaned namespaces in case previous run was interrupted.
    : ${E2E_CLEAN_START:="true"}
    : ${PROJECT:="kubernetes-jenkins"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GKE_REQUIRED_SKIP_TESTS[@]:+${GKE_REQUIRED_SKIP_TESTS[@]}} \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
          ${GCE_SOAK_CONTINUOUS_SKIP_TESTS[@]:+${GCE_SOAK_CONTINUOUS_SKIP_TESTS[@]}} \
          )"}
    ;;

  # kubernetes-upgrade-gke

  kubernetes-upgrade-gke-step1-deploy)
    configure_upgrade_step 'release/latest' 'ci/latest' 'gke-upgrade' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-step2-upgrade-master)
    configure_upgrade_step 'release/latest' 'ci/latest' 'gke-upgrade' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-step3-e2e-old)
    configure_upgrade_step 'release/latest' 'ci/latest' 'gke-upgrade' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-step4-upgrade-cluster)
    configure_upgrade_step 'release/latest' 'ci/latest' 'gke-upgrade' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-step5-e2e-old)
    configure_upgrade_step 'release/latest' 'ci/latest' 'gke-upgrade' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-step6-e2e-new)
    configure_upgrade_step 'release/latest' 'ci/latest' 'gke-upgrade' 'kubernetes-jenkins-gke-upgrade'
    ;;

  # kubernetes-upgrade-gke-stable-latest
  #
  # This suite:
  #
  # 1. launches a cluster at release/stable,
  # 2. upgrades the master to release/latest,
  # 3. runs release/stable e2es,
  # 4. upgrades the rest of the cluster,
  # 5. runs release/stable e2es again, then
  # 6. runs release/latest e2es and tears down the cluster.

  kubernetes-upgrade-stable-latest-gke-step1-deploy)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="gke-upgrade-stable-latest"}
    : ${E2E_NETWORK:="gke-upgrade-stable-latest"}
    : ${JENKINS_PUBLISHED_VERSION:="release/stable"}
    : ${E2E_SET_CLUSTER_API_VERSION:=y}
    : ${PROJECT:="k8s-jkns-upgrade-fixed-1"}
    : ${E2E_UP:="true"}
    : ${E2E_TEST:="false"}
    : ${E2E_DOWN:="false"}
    ;;

  kubernetes-upgrade-stable-latest-gke-step2-upgrade-master)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="gke-upgrade-stable-latest"}
    : ${E2E_NETWORK:="gke-upgrade-stable-latest"}
    : ${E2E_OPT:="--check_version_skew=false"}
    # Use upgrade logic of version we're upgrading to.
    : ${JENKINS_PUBLISHED_VERSION:="release/latest"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${PROJECT:="k8s-jkns-upgrade-fixed-1"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=Cluster\sUpgrade.*upgrade-master --upgrade-target=release/latest"}
    ;;

  kubernetes-upgrade-stable-latest-gke-step3-e2e-old)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="gke-upgrade-stable-latest"}
    : ${E2E_NETWORK:="gke-upgrade-stable-latest"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    # Run old e2es
    : ${JENKINS_PUBLISHED_VERSION:="release/stable"}
    : ${PROJECT:="k8s-jkns-upgrade-fixed-1"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GKE_DEFAULT_SKIP_TESTS[@]:+${GKE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
          )"}
    ;;

  kubernetes-upgrade-stable-latest-gke-step4-upgrade-cluster)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="gke-upgrade-stable-latest"}
    : ${E2E_NETWORK:="gke-upgrade-stable-latest"}
    : ${E2E_OPT:="--check_version_skew=false"}
    # Use upgrade logic of version we're upgrading to.
    : ${JENKINS_PUBLISHED_VERSION:="release/latest"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${PROJECT:="k8s-jkns-upgrade-fixed-1"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=Cluster\sUpgrade.*upgrade-cluster --upgrade-target=release/latest"}
    ;;

  kubernetes-upgrade-stable-latest-gke-step5-e2e-old)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="gke-upgrade-stable-latest"}
    : ${E2E_NETWORK:="gke-upgrade-stable-latest"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    # Run old e2es
    : ${JENKINS_PUBLISHED_VERSION:="release/stable"}
    : ${PROJECT:="k8s-jkns-upgrade-fixed-1"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GKE_DEFAULT_SKIP_TESTS[@]:+${GKE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
          )"}
    ;;

  kubernetes-upgrade-stable-latest-gke-step6-e2e-new)
    : ${DOGFOOD_GCLOUD:="true"}
    : ${GKE_API_ENDPOINT:="https://test-container.sandbox.googleapis.com/"}
    : ${E2E_CLUSTER_NAME:="gke-upgrade-stable-latest"}
    : ${E2E_NETWORK:="gke-upgrade-stable-latest"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${JENKINS_PUBLISHED_VERSION:="release/latest"}
    : ${PROJECT:="k8s-jkns-upgrade-fixed-1"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="true"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GKE_DEFAULT_SKIP_TESTS[@]:+${GKE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GKE_FLAKY_TESTS[@]:+${GKE_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    ;;

  # kubernetes-upgrade-gke-1.0-master
  #
  # Test upgrades from the latest release-1.0 build to the latest master build.
  #
  # Configurations for step1, step3, and step5 live in the release-1.0 branch.

  kubernetes-upgrade-gke-1.0-master-step2-upgrade-master)
    configure_upgrade_step 'configured-in-release-1.0' 'ci/latest' 'upgrade-gke-1-0-master' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-1.0-master-step4-upgrade-cluster)
    configure_upgrade_step 'configured-in-release-1.0' 'ci/latest' 'upgrade-gke-1-0-master' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-1.0-master-step6-e2e-new)
    configure_upgrade_step 'configured-in-release-1.0' 'ci/latest' 'upgrade-gke-1-0-master' 'kubernetes-jenkins-gke-upgrade'
    ;;

  # kubernetes-upgrade-gke-1.1-master
  #
  # Test upgrades from the latest release-1.1 build to the latest master build.
  #
  # Configurations for step1, step3, and step5 live in the release-1.1 branch.

  kubernetes-upgrade-gke-1.1-master-step2-upgrade-master)
    configure_upgrade_step 'configured-in-release-1.1' 'ci/latest' 'upgrade-gke-1-1-master' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-1.1-master-step4-upgrade-cluster)
    configure_upgrade_step 'configured-in-release-1.1' 'ci/latest' 'upgrade-gke-1-1-master' 'kubernetes-jenkins-gke-upgrade'
    ;;

  kubernetes-upgrade-gke-1.1-master-step6-e2e-new)
    configure_upgrade_step 'configured-in-release-1.1' 'ci/latest' 'upgrade-gke-1-1-master' 'kubernetes-jenkins-gke-upgrade'
    ;;

  # kubernetes-upgrade-gce
  #
  # This suite:
  #
  # 1. launches a cluster at release/latest,
  # 2. upgrades the master to ci/latest,
  # 3. runs release/latest e2es,
  # 4. upgrades the rest of the cluster,
  # 5. runs release/latest e2es again, then
  # 6. runs ci/latest e2es and tears down the cluster.

  kubernetes-upgrade-gce-step1-deploy)
    : ${E2E_CLUSTER_NAME:="gce-upgrade"}
    : ${E2E_NETWORK:="gce-upgrade"}
    : ${JENKINS_PUBLISHED_VERSION:="release/latest"}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="true"}
    : ${E2E_TEST:="false"}
    : ${E2E_DOWN:="false"}
    : ${NUM_NODES:=5}
    ;;

  kubernetes-upgrade-gce-step2-upgrade-master)
    : ${E2E_CLUSTER_NAME:="gce-upgrade"}
    : ${E2E_NETWORK:="gce-upgrade"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=Cluster\sUpgrade.*upgrade-master"}
    : ${NUM_NODES:=5}
    : ${KUBE_ENABLE_DEPLOYMENTS:=true}
    : ${KUBE_ENABLE_DAEMONSETS:=true}
    ;;

  kubernetes-upgrade-gce-step3-e2e-old)
    : ${E2E_CLUSTER_NAME:="gce-upgrade"}
    : ${E2E_NETWORK:="gce-upgrade"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${JENKINS_USE_RELEASE_TARS:=y}
    : ${JENKINS_FORCE_GET_TARS:=y}
    # Run release/latest e2es
    : ${JENKINS_PUBLISHED_VERSION:="release/latest"}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          )"}
    : ${NUM_NODES:=5}
    ;;

  kubernetes-upgrade-gce-step4-upgrade-cluster)
    : ${E2E_CLUSTER_NAME:="gce-upgrade"}
    : ${E2E_NETWORK:="gce-upgrade"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=Cluster\sUpgrade.*upgrade-cluster"}
    : ${NUM_NODES:=5}
    : ${KUBE_ENABLE_DEPLOYMENTS:=true}
    : ${KUBE_ENABLE_DAEMONSETS:=true}
    ;;

  kubernetes-upgrade-gce-step5-e2e-old)
    : ${E2E_CLUSTER_NAME:="gce-upgrade"}
    : ${E2E_NETWORK:="gce-upgrade"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    # Run release/latest e2es
    : ${JENKINS_PUBLISHED_VERSION:="release/latest"}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          )"}
    : ${NUM_NODES:=5}
    ;;

  kubernetes-upgrade-gce-step6-e2e-new)
    : ${E2E_CLUSTER_NAME:="gce-upgrade"}
    : ${E2E_NETWORK:="gce-upgrade"}
    # TODO(15011): these really shouldn't be (very) version skewed, but because
    # we have to get ci/latest again, it could get slightly out of whack.
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="true"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    : ${NUM_NODES:=5}
    ;;

  # kubernetes-upgrade-gce-1.0-current-release
  #
  # This suite:
  #
  # 1. launches a cluster at ci/latest-1.0,
  # 2. upgrades the master to CURRENT_RELEASE_PUBLISHED_VERSION
  # 3. runs ci/latest-1.0 e2es,
  # 4. upgrades the rest of the cluster,
  # 5. runs ci/latest-1.0 e2es again, then
  # 6. runs CURRENT_RELEASE_PUBLISHED_VERSION e2es and tears down the cluster.

  kubernetes-upgrade-1.0-current-release-gce-step1-deploy)
    : ${E2E_CLUSTER_NAME:="gce-upgrade-1-0"}
    : ${E2E_NETWORK:="gce-upgrade-1-0"}
    : ${JENKINS_PUBLISHED_VERSION:="ci/latest-1.0"}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="true"}
    : ${E2E_TEST:="false"}
    : ${E2E_DOWN:="false"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-upgrade-1-0"}
    : ${NUM_NODES:=5}
    ;;

  kubernetes-upgrade-1.0-current-release-gce-step2-upgrade-master)
    : ${E2E_CLUSTER_NAME:="gce-upgrade-1-0"}
    : ${E2E_NETWORK:="gce-upgrade-1-0"}
    : ${E2E_OPT:="--check_version_skew=false"}
    # Use upgrade logic of version we're upgrading to.
    : ${JENKINS_PUBLISHED_VERSION:="${CURRENT_RELEASE_PUBLISHED_VERSION}"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=Cluster\sUpgrade.*upgrade-master --upgrade-target=${CURRENT_RELEASE_PUBLISHED_VERSION}"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-upgrade-1-0"}
    : ${NUM_NODES:=5}
    : ${KUBE_ENABLE_DEPLOYMENTS:=true}
    : ${KUBE_ENABLE_DAEMONSETS:=true}
    ;;

  kubernetes-upgrade-1.0-current-release-gce-step3-e2e-old)
    : ${E2E_CLUSTER_NAME:="gce-upgrade-1-0"}
    : ${E2E_NETWORK:="gce-upgrade-1-0"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    # Run old e2es
    : ${JENKINS_PUBLISHED_VERSION:="ci/latest-1.0"}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-upgrade-1-0"}
    : ${NUM_NODES:=5}
    ;;

  kubernetes-upgrade-1.0-current-release-gce-step4-upgrade-cluster)
    : ${E2E_CLUSTER_NAME:="gce-upgrade-1-0"}
    : ${E2E_NETWORK:="gce-upgrade-1-0"}
    : ${E2E_OPT:="--check_version_skew=false"}
    # Use upgrade logic of version we're upgrading to.
    : ${JENKINS_PUBLISHED_VERSION:="${CURRENT_RELEASE_PUBLISHED_VERSION}"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.focus=Cluster\sUpgrade.*upgrade-cluster --upgrade-target=${CURRENT_RELEASE_PUBLISHED_VERSION}"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-upgrade-1-0"}
    : ${NUM_NODES:=5}
    : ${KUBE_ENABLE_DEPLOYMENTS:=true}
    : ${KUBE_ENABLE_DAEMONSETS:=true}
    ;;

  kubernetes-upgrade-1.0-current-release-gce-step5-e2e-old)
    : ${E2E_CLUSTER_NAME:="gce-upgrade-1-0"}
    : ${E2E_NETWORK:="gce-upgrade-1-0"}
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    # Run old e2es
    : ${JENKINS_PUBLISHED_VERSION:="ci/latest-1.0"}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="false"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-upgrade-1-0"}
    : ${NUM_NODES:=5}
    ;;

  kubernetes-upgrade-1.0-current-release-gce-step6-e2e-new)
    : ${E2E_CLUSTER_NAME:="gce-upgrade-1-0"}
    : ${E2E_NETWORK:="gce-upgrade-1-0"}
    # TODO(15011): these really shouldn't be (very) version skewed, but because
    # we have to get CURRENT_RELEASE_PUBLISHED_VERSION again, it could get slightly out of whack.
    : ${E2E_OPT:="--check_version_skew=false"}
    : ${JENKINS_FORCE_GET_TARS:=y}
    : ${JENKINS_PUBLISHED_VERSION:="${CURRENT_RELEASE_PUBLISHED_VERSION}"}
    : ${PROJECT:="k8s-jkns-gce-upgrade"}
    : ${E2E_UP:="false"}
    : ${E2E_TEST:="true"}
    : ${E2E_DOWN:="true"}
    : ${GINKGO_TEST_ARGS:="--ginkgo.skip=$(join_regex_allow_empty \
          ${GCE_DEFAULT_SKIP_TESTS[@]:+${GCE_DEFAULT_SKIP_TESTS[@]}} \
          ${GCE_FLAKY_TESTS[@]:+${GCE_FLAKY_TESTS[@]}} \
          ${GCE_SLOW_TESTS[@]:+${GCE_SLOW_TESTS[@]}} \
          )"}
    : ${KUBE_GCE_INSTANCE_PREFIX:="e2e-upgrade-1-0"}
    : ${NUM_NODES:=5}
    ;;

  # Run Kubemark test on a fake 100 node cluster to have a comparison
  # to the real results from scalability suite
  kubernetes-kubemark-gce)
    : ${E2E_CLUSTER_NAME:="kubernetes-kubemark"}
    : ${E2E_NETWORK:="kubernetes-kubemark"}
    : ${PROJECT:="k8s-jenkins-kubemark"}
    : ${E2E_UP:="true"}
    : ${E2E_DOWN:="true"}
    : ${E2E_TEST:="false"}
    : ${USE_KUBEMARK:="true"}
    # Override defaults to be indpendent from GCE defaults and set kubemark parameters
    KUBE_GCE_INSTANCE_PREFIX="kubemark100"
    NUM_NODES="10"
    MASTER_SIZE="n1-standard-2"
    NODE_SIZE="n1-standard-1"
    KUBEMARK_MASTER_SIZE="n1-standard-4"
    KUBEMARK_NUM_NODES="100"
    ;;

  # Run Kubemark test on a fake 500 node cluster to test for regressions on
  # bigger clusters
  kubernetes-kubemark-500-gce)
    : ${E2E_CLUSTER_NAME:="kubernetes-kubemark-500"}
    : ${E2E_NETWORK:="kubernetes-kubemark-500"}
    : ${PROJECT:="k8s-jenkins-kubemark"}
    : ${E2E_UP:="true"}
    : ${E2E_DOWN:="true"}
    : ${E2E_TEST:="false"}
    : ${USE_KUBEMARK:="true"}
    # Override defaults to be indpendent from GCE defaults and set kubemark parameters
    NUM_NODES="6"
    MASTER_SIZE="n1-standard-4"
    NODE_SIZE="n1-standard-8"
    KUBE_GCE_INSTANCE_PREFIX="kubemark500"
    E2E_ZONE="asia-east1-a"
    KUBEMARK_MASTER_SIZE="n1-standard-16"
    KUBEMARK_NUM_NODES="500"
    ;;

  # Run big Kubemark test, this currently means a 1000 node cluster and 16 core master
  kubernetes-kubemark-gce-scale)
    : ${E2E_CLUSTER_NAME:="kubernetes-kubemark-scale"}
    : ${E2E_NETWORK:="kubernetes-kubemark-scale"}
    : ${PROJECT:="kubernetes-scale"}
    : ${E2E_UP:="true"}
    : ${E2E_DOWN:="true"}
    : ${E2E_TEST:="false"}
    : ${USE_KUBEMARK:="true"}
    # Override defaults to be indpendent from GCE defaults and set kubemark parameters
    # We need 11 so that we won't hit max-pods limit (set to 100). TODO: do it in a nicer way.
    NUM_NODES="11"
    MASTER_SIZE="n1-standard-4"
    NODE_SIZE="n1-standard-8"   # Note: can fit about 17 hollow nodes per core
    #                                     so NUM_NODES x cores_per_node should
    #                                     be set accordingly.
    KUBE_GCE_INSTANCE_PREFIX="kubemark1000"
    E2E_ZONE="asia-east1-a"
    KUBEMARK_MASTER_SIZE="n1-standard-16"
    KUBEMARK_NUM_NODES="1000"
    ;;
esac

# AWS variables
export KUBE_AWS_INSTANCE_PREFIX=${E2E_CLUSTER_NAME}
export KUBE_AWS_ZONE=${E2E_ZONE}

# GCE variables
export INSTANCE_PREFIX=${E2E_CLUSTER_NAME}
export KUBE_GCE_ZONE=${E2E_ZONE}
export KUBE_GCE_NETWORK=${E2E_NETWORK}
export KUBE_GCE_INSTANCE_PREFIX=${KUBE_GCE_INSTANCE_PREFIX:-}
export KUBE_GCS_STAGING_PATH_SUFFIX=${KUBE_GCS_STAGING_PATH_SUFFIX:-}
export KUBE_GCE_NODE_PROJECT=${KUBE_GCE_NODE_PROJECT:-}
export KUBE_GCE_NODE_IMAGE=${KUBE_GCE_NODE_IMAGE:-}
export KUBE_OS_DISTRIBUTION=${KUBE_OS_DISTRIBUTION:-}

# GKE variables
export CLUSTER_NAME=${E2E_CLUSTER_NAME}
export ZONE=${E2E_ZONE}
export KUBE_GKE_NETWORK=${E2E_NETWORK}
export E2E_SET_CLUSTER_API_VERSION=${E2E_SET_CLUSTER_API_VERSION:-}
export DOGFOOD_GCLOUD=${DOGFOOD_GCLOUD:-}
export CMD_GROUP=${CMD_GROUP:-}
export MACHINE_TYPE=${NODE_SIZE:-}  # GKE scripts use MACHINE_TYPE for the node vm size

if [[ ! -z "${GKE_API_ENDPOINT:-}" ]]; then
  export CLOUDSDK_API_ENDPOINT_OVERRIDES_CONTAINER=${GKE_API_ENDPOINT}
fi

# Shared cluster variables
export E2E_MIN_STARTUP_PODS=${E2E_MIN_STARTUP_PODS:-}
export KUBE_ENABLE_CLUSTER_MONITORING=${ENABLE_CLUSTER_MONITORING:-}
export KUBE_ENABLE_CLUSTER_REGISTRY=${ENABLE_CLUSTER_REGISTRY:-}
export KUBE_ENABLE_HORIZONTAL_POD_AUTOSCALER=${ENABLE_HORIZONTAL_POD_AUTOSCALER:-}
export KUBE_ENABLE_DEPLOYMENTS=${ENABLE_DEPLOYMENTS:-}
export KUBE_ENABLE_EXPERIMENTAL_API=${ENABLE_EXPERIMENTAL_API:-}
export MASTER_SIZE=${MASTER_SIZE:-}
export NODE_SIZE=${NODE_SIZE:-}
export NODE_DISK_SIZE=${NODE_DISK_SIZE:-}
export NUM_NODES=${NUM_NODES:-}
export TEST_CLUSTER_LOG_LEVEL=${TEST_CLUSTER_LOG_LEVEL:-}
export TEST_CLUSTER_RESYNC_PERIOD=${TEST_CLUSTER_RESYNC_PERIOD:-}
export PROJECT=${PROJECT:-}
export JENKINS_EXPLICIT_VERSION=${JENKINS_EXPLICIT_VERSION:-}
export JENKINS_PUBLISHED_VERSION=${JENKINS_PUBLISHED_VERSION:-'ci/latest'}

export KUBE_ADMISSION_CONTROL=${ADMISSION_CONTROL:-}

export KUBERNETES_PROVIDER=${KUBERNETES_PROVIDER}
export PATH=${PATH}:/usr/local/go/bin
export KUBE_SKIP_CONFIRMATIONS=y

# E2E Control Variables
export E2E_UP="${E2E_UP:-true}"
export E2E_TEST="${E2E_TEST:-true}"
export E2E_DOWN="${E2E_DOWN:-true}"
export E2E_CLEAN_START="${E2E_CLEAN_START:-}"
# Used by hack/ginkgo-e2e.sh to enable ginkgo's parallel test runner.
export GINKGO_PARALLEL=${GINKGO_PARALLEL:-}

echo "--------------------------------------------------------------------------------"
echo "Test Environment:"
printenv | sort
echo "--------------------------------------------------------------------------------"

# We get the Kubernetes tarballs on either cluster creation or when we want to
# replace existing ones in a multi-step job (e.g. a cluster upgrade).
if [[ "${E2E_UP,,}" == "true" || "${JENKINS_FORCE_GET_TARS:-}" =~ ^[yY]$ ]]; then
    if [[ ${KUBE_RUN_FROM_OUTPUT:-} =~ ^[yY]$ ]]; then
        echo "Found KUBE_RUN_FROM_OUTPUT=y; will use binaries from _output"
        cp _output/release-tars/kubernetes*.tar.gz .
    else
        echo "Pulling binaries from GCS"
        # In a multi-step job, clean up just the kubernetes build files.
        # Otherwise, we want a completely empty directory.
        if [[ "${JENKINS_FORCE_GET_TARS:-}" =~ ^[yY]$ ]]; then
            rm -rf kubernetes*
        elif [[ $(find . | wc -l) != 1 ]]; then
            echo $PWD not empty, bailing!
            exit 1
        fi

        # Tell kube-up.sh to skip the update, it doesn't lock. An internal
        # gcloud bug can cause racing component updates to stomp on each
        # other.
        export KUBE_SKIP_UPDATE=y
        (
          # ----------- WARNING! DO NOT TOUCH THIS CODE -----------
          #
          # The purpose of this block is to ensure that only one job attempts to
          # update gcloud at a time.
          #
          # PLEASE DO NOT TOUCH THIS CODE unless you are certain you understand
          # implications. Please cc jlowdermilk@ or brendandburns@ on changes.

          # If jenkins was recently restarted and jobs are failing with
          #
          # flock: 9: Permission denied
          #
          # ssh into the jenkins master and run
          # $ sudo chown jenkins:jenkins /var/run/lock/gcloud-components.lock
          #
          # Note, flock -n would prevent parallel runs from having to wait
          # here, but because we've set -o errexit, the err gets caught
          # despite running in a subshell. If a run has to wait, the subsequent
          # component update commands will be no-op, so no added delay.
          flock -x -w 60 9
          # We do NOT want to run gcloud components update under sudo, as that causes
          # the gcloud files to get chown'd by root, which makes them undeletable in
          # the case where we are installing gcloud under the workspace (e.g. for gke-ci
          # and friends). If we can't cleanup old workspaces, jenkins runs out of disk.
          #
          # If the update commands are failing with permission denied, ssh into
          # the jenkins master and run
          #
          # $ sudo chown -R jenkins:jenkins /usr/local/share/google/google-cloud-sdk
          gcloud components update -q || true
          gcloud components update alpha -q || true
          gcloud components update beta -q || true
        ) 9>/var/run/lock/gcloud-components.lock

        if [[ ! -z ${JENKINS_EXPLICIT_VERSION:-} ]]; then
            # Use an explicit pinned version like "ci/v0.10.0-101-g6c814c4" or
            # "release/v0.19.1"
            IFS='/' read -a varr <<< "${JENKINS_EXPLICIT_VERSION}"
            bucket="${varr[0]}"
            githash="${varr[1]}"
            echo "Using explicit version $bucket/$githash"
        elif [[ ${JENKINS_USE_SERVER_VERSION:-}  =~ ^[yY]$ ]]; then
            # for GKE we can use server default version.
            bucket="release"
            msg=$(gcloud ${CMD_GROUP} container get-server-config --project=${PROJECT} --zone=${ZONE} | grep defaultClusterVersion)
            # msg will look like "defaultClusterVersion: 1.0.1". Strip
            # everything up to, including ": "
            githash="v${msg##*: }"
            echo "Using server version $bucket/$githash"
        else  # use JENKINS_PUBLISHED_VERSION
            # Use a published version like "ci/latest" (default),
            # "release/latest", "release/latest-1", or "release/stable"
            IFS='/' read -a varr <<< "${JENKINS_PUBLISHED_VERSION}"
            bucket="${varr[0]}"
            githash=$(gsutil cat gs://kubernetes-release/${JENKINS_PUBLISHED_VERSION}.txt)
            echo "Using published version $bucket/$githash (from ${JENKINS_PUBLISHED_VERSION})"
        fi
        # At this point, we want to have the following vars set:
        # - bucket
        # - githash
        gsutil -m cp gs://kubernetes-release/${bucket}/${githash}/kubernetes.tar.gz gs://kubernetes-release/${bucket}/${githash}/kubernetes-test.tar.gz .

        # Set by GKE-CI to change the CLUSTER_API_VERSION to the git version
        if [[ ! -z ${E2E_SET_CLUSTER_API_VERSION:-} ]]; then
            export CLUSTER_API_VERSION=$(echo ${githash} | cut -c 2-)
        fi
    fi

    if [[ ! "${CIRCLECI:-}" == "true" ]]; then
        # Copy GCE keys so we don't keep cycling them.
        # To set this up, you must know the <project>, <zone>, and <instance>
        # on which your jenkins jobs are running. Then do:
        #
        # # SSH from your computer into the instance.
        # $ gcloud compute ssh --project="<prj>" ssh --zone="<zone>" <instance>
        #
        # # Generate a key by ssh'ing from the instance into itself, then exit.
        # $ gcloud compute ssh --project="<prj>" ssh --zone="<zone>" <instance>
        # $ ^D
        #
        # # Copy the keys to the desired location (e.g. /var/lib/jenkins/gce_keys/).
        # $ sudo mkdir -p /var/lib/jenkins/gce_keys/
        # $ sudo cp ~/.ssh/google_compute_engine /var/lib/jenkins/gce_keys/
        # $ sudo cp ~/.ssh/google_compute_engine.pub /var/lib/jenkins/gce_keys/
        #
        # # Move the permissions for the keys to Jenkins.
        # $ sudo chown -R jenkins /var/lib/jenkins/gce_keys/
        # $ sudo chgrp -R jenkins /var/lib/jenkins/gce_keys/
        if [[ "${KUBERNETES_PROVIDER}" == "aws" ]]; then
            echo "Skipping SSH key copying for AWS"
        else
            mkdir -p ${WORKSPACE}/.ssh/
            cp /var/lib/jenkins/gce_keys/google_compute_engine ${WORKSPACE}/.ssh/
            cp /var/lib/jenkins/gce_keys/google_compute_engine.pub ${WORKSPACE}/.ssh/
        fi
    fi

    md5sum kubernetes*.tar.gz
    tar -xzf kubernetes.tar.gz
    tar -xzf kubernetes-test.tar.gz
fi

cd kubernetes

# Have cmd/e2e run by goe2e.sh generate JUnit report in ${WORKSPACE}/junit*.xml
ARTIFACTS=${WORKSPACE}/_artifacts
mkdir -p ${ARTIFACTS}
export E2E_REPORT_DIR=${ARTIFACTS}
declare -r gcp_list_resources_script="./cluster/gce/list-resources.sh"
declare -r gcp_resources_before="${ARTIFACTS}/gcp-resources-before.txt"
declare -r gcp_resources_cluster_up="${ARTIFACTS}/gcp-resources-cluster-up.txt"
declare -r gcp_resources_after="${ARTIFACTS}/gcp-resources-after.txt"
# TODO(15492): figure out some way to run this script even if it doesn't exist
# in the Kubernetes tarball.
if [[ ( ${KUBERNETES_PROVIDER} == "gce" || ${KUBERNETES_PROVIDER} == "gke" ) && -x "${gcp_list_resources_script}" ]]; then
  gcp_list_resources="true"
else
  gcp_list_resources="false"
fi

### Pre Set Up ###
# Install gcloud from a custom path if provided. Used to test GKE with gcloud
# at HEAD, release candidate.
if [[ ! -z "${CLOUDSDK_BUCKET:-}" ]]; then
    gsutil -m cp -r "${CLOUDSDK_BUCKET}" ~
    mv ~/$(basename "${CLOUDSDK_BUCKET}") ~/repo
    mkdir ~/cloudsdk
    tar zvxf ~/repo/google-cloud-sdk.tar.gz -C ~/cloudsdk
    export CLOUDSDK_CORE_DISABLE_PROMPTS=1
    export CLOUDSDK_COMPONENT_MANAGER_SNAPSHOT_URL=file://${HOME}/repo/components-2.json
    ~/cloudsdk/google-cloud-sdk/install.sh --disable-installation-options --bash-completion=false --path-update=false --usage-reporting=false
    export PATH=${HOME}/cloudsdk/google-cloud-sdk/bin:${PATH}
    export CLOUDSDK_CONFIG=/var/lib/jenkins/.config/gcloud
fi

### Set up ###
if [[ "${E2E_UP,,}" == "true" ]]; then
    go run ./hack/e2e.go ${E2E_OPT} -v --down
fi
if [[ "${gcp_list_resources}" == "true" ]]; then
  ${gcp_list_resources_script} > "${gcp_resources_before}"
fi
if [[ "${E2E_UP,,}" == "true" ]]; then
    go run ./hack/e2e.go ${E2E_OPT} -v --up
    go run ./hack/e2e.go -v --ctl="version --match-server-version=false"
    if [[ "${gcp_list_resources}" == "true" ]]; then
      ${gcp_list_resources_script} > "${gcp_resources_cluster_up}"
    fi
fi

### Run tests ###
# Jenkins will look at the junit*.xml files for test failures, so don't exit
# with a nonzero error code if it was only tests that failed.
if [[ "${E2E_TEST,,}" == "true" ]]; then
    go run ./hack/e2e.go ${E2E_OPT} -v --test --test_args="${GINKGO_TEST_ARGS}" && exitcode=0 || exitcode=$?
    if [[ "${E2E_PUBLISH_GREEN_VERSION:-}" == "true" && ${exitcode} == 0 && -n ${githash:-} ]]; then
        echo "publish githash to ci/latest-green.txt: ${githash}"
        echo "${githash}" > ${WORKSPACE}/githash.txt
        gsutil cp ${WORKSPACE}/githash.txt gs://kubernetes-release/ci/latest-green.txt
    fi
fi

### Start Kubemark ###
if [[ "${USE_KUBEMARK:-}" == "true" ]]; then
  export RUN_FROM_DISTRO=true
  NUM_NODES_BKP=${NUM_NODES}
  MASTER_SIZE_BKP=${MASTER_SIZE}
  ./test/kubemark/stop-kubemark.sh
  NUM_NODES=${KUBEMARK_NUM_NODES:-$NUM_NODES}
  MASTER_SIZE=${KUBEMARK_MASTER_SIZE:-$MASTER_SIZE}
  ./test/kubemark/start-kubemark.sh
  ./test/kubemark/run-e2e-tests.sh --ginkgo.focus="should\sallow\sstarting\s30\spods\sper\snode" --delete-namespace="false" --gather-resource-usage="false"
  ./test/kubemark/stop-kubemark.sh
  NUM_NODES=${NUM_NODES_BKP}
  MASTER_SIZE=${MASTER_SIZE_BKP}
  unset RUN_FROM_DISTRO
  unset NUM_NODES_BKP
  unset MASTER_SIZE_BKP
fi

### Clean up ###
if [[ "${E2E_DOWN,,}" == "true" ]]; then
    # Sleep before deleting the cluster to give the controller manager time to
    # delete any cloudprovider resources still around from the last test.
    # This is calibrated to allow enough time for 3 attempts to delete the
    # resources. Each attempt is allocated 5 seconds for requests to the
    # cloudprovider plus the processingRetryInterval from servicecontroller.go
    # for the wait between attempts.
    sleep 30
    go run ./hack/e2e.go ${E2E_OPT} -v --down
fi
if [[ "${gcp_list_resources}" == "true" ]]; then
  ${gcp_list_resources_script} > "${gcp_resources_after}"
fi

# Compare resources if either the cluster was
# * started and destroyed (normal e2e)
# * neither started nor destroyed (soak test)
if [[ "${E2E_UP:-}" == "${E2E_DOWN:-}" && -f "${gcp_resources_before}" && -f "${gcp_resources_after}" ]]; then
  if ! diff -sw -U0 -F'^\[.*\]$' "${gcp_resources_before}" "${gcp_resources_after}" && [[ "${FAIL_ON_GCP_RESOURCE_LEAK:-}" == "true" ]]; then
    echo "!!! FAIL: Google Cloud Platform resources leaked while running tests!"
    exit 1
  fi
fi
