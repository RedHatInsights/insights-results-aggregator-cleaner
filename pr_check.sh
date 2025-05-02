#!/bin/bash

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ccx-data-pipeline"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="insights-aggregator-cleaner"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/insights-results-aggregator-cleaner"
COMPONENTS="ccx-insights-results insights-aggregator-cleaner dvo-writer"  # space-separated list of components to laod
COMPONENTS_W_RESOURCES="insights-aggregator-cleaner"  # component to keep
CACHE_FROM_LATEST_IMAGE="false"

export IQE_PLUGINS="ccx"
export IQE_MARKER_EXPRESSION=""
# Workaround: There are no cleaner specific integration tests. Check that the service loads and iqe plugin works.
export IQE_FILTER_EXPRESSION="test_plugin_accessible"
export IQE_REQUIREMENTS_PRIORITY=""
export IQE_TEST_IMPORTANCE=""
export IQE_CJI_TIMEOUT="30m"
export IQE_ENV="ephemeral"
export IQE_ENV_VARS="DYNACONF_USER_PROVIDER__rbac_enabled=false"
# Set the correct images for pull requests.
# pr_check in pull requests still uses the old cloudservices images
EXTRA_DEPLOY_ARGS="--set-parameter insights-aggregator-cleaner/IMAGE=quay.io/cloudservices/insights-results-aggregator-cleaner"


function build_image() {
    source $CICD_ROOT/build.sh
}

function deploy_ephemeral() {
    # shellcheck disable=SC2317
    source $CICD_ROOT/deploy_ephemeral_env.sh
}

function run_smoke_tests() {
    # shellcheck disable=SC2317
    source $CICD_ROOT/cji_smoke_test.sh
    # shellcheck disable=SC2317
    source $CICD_ROOT/post_test_results.sh  # publish results in Ibutsu
}


# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh
echo "creating PR image"
# shellcheck disable=SC2317
build_image

echo "deploying to ephemeral"
# shellcheck disable=SC2317
deploy_ephemeral

echo "running PR smoke tests"
# shellcheck disable=SC2317
run_smoke_tests
