# Copyright © 2021 Pavel Tisnovsky, Red Hat, Inc.
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

import subprocess


test_output = "test"


def process_cleaner_output(context, out, return_code):
    """Process cleaner output."""
    # interact with the process:
    # read data from stdout and stderr, until end-of-file is reached
    stdout, stderr = out.communicate()

    # basic checks
    assert stderr is None, "Error during check"
    assert stdout is not None, "No output from cleaner"

    # assert stdout is None, "{}".format(stdout)
    # assert stderr is None, "{}".format(stderr)
    assert out.returncode == 0 or out.returncode == return_code, \
        "Return code is {}".format(out.returncode)

    output = stdout.decode('utf-8').split("\n")

    assert output is not None
    context.output = output
    context.stdout = stdout
    context.stderr = stderr


@when(u"I run the cleaner to display all records older than {age}")
def run_cleaner_for_older_records(context, age):
    """Start the cleaner to retrieve list of older records."""
    out = subprocess.Popen(["insights-results-aggregator-cleaner", "--output", test_output,
                            "--max-age", age],
                           stdout=subprocess.PIPE,
                           stderr=subprocess.STDOUT)

    process_cleaner_output(context, out, 0)


@when(u"I run the cleaner with the {flag} command line flag")
def run_cleaner_with_flag(context, flag):
    """Start the cleaner with given command-line flag."""
    out = subprocess.Popen(["insights-results-aggregator-cleaner", flag],
                           stdout=subprocess.PIPE,
                           stderr=subprocess.STDOUT)

    process_cleaner_output(context, out, 2)


@then(u"I should see help messages displayed on standard output")
def check_help_from_cleaner(context):
    """Check if help is displayed by cleaner."""
    expected_output = """
Usage of insights-results-aggregator-cleaner:
  -cleanup
        perform database cleanup
  -clusters string
        list of clusters to cleanup
  -fill-in-db
        fill-in database by test data
  -max-age string
        max age for displaying old records
  -output string
        filename for old cluster listing
  -summary
        print summary table after cleanup
  -version
        show cleaner version"""

    stdout = context.stdout.decode("utf-8").replace("\t", "    ")
    assert stdout.strip() == expected_output.strip(), "{} != {}".format(stdout, expected_output)


@then(u"I should see version info displayed on standard output")
def check_version_from_cleaner(context):
    """Check if version info is displayed by cleaner."""
    assert "Insights Results Aggregator Cleaner version 1.0" in context.output 


@then(u"I should see empty list of records")
def check_empty_list_of_records(context):
    """Check if the cleaner displays empty list of records."""
    with open(test_output, "r") as fin:
        content = fin.read()
        assert content == "", "expecting empty list of clusters"
