# Benchmarking Stack for AWS Lambda Runtimes
An all-in-one benchmarking stack for AWS Lambda supporting Node.js, Python, Go and Rust. 

Important note: This application utilizes various AWS services, and there may be costs associated with these services beyond the usage covered by the Free Tier. It's crucial to regularly monitor and manage your AWS usage to avoid unexpected charges. You can set up billing alerts, track your usage with AWS Cost Explorer, and utilize various cost optimization tools and best practices to help manage your AWS costs effectively.

## Requirements

* [Create an AWS account](https://portal.aws.amazon.com/gp/aws/developer/registration/index.html) if you do not already have one and log in. The IAM user that you use must have sufficient permissions to make necessary AWS service calls and manage AWS resources.
* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) installed and configured
* [Git Installed](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* [AWS Serverless Application Model](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) (AWS SAM) installed
* [Rust](https://www.rust-lang.org/) 1.56.0 or higher
* [Go](https://go.dev/dl/) (`1.16` or above) installed


## Deployment Instructions

1. Create a new directory, navigate to that directory in a terminal and clone the GitHub repository:
    ``` 
    git clone https://github.com/nigelcss/bmlambda
    ```
1. Change directory to the project directory:
    ```
    cd bmlambda
    ```
1. Install dependencies and build:
    ```
    make build
    ```
1. From the command line, use AWS SAM to deploy the AWS resources for the pattern as specified in the template.yml file:
    ```
    make deploy
    ```
1. During the prompts:
    * Enter a stack name
    * Enter the desired AWS Region
    * Allow SAM CLI to create IAM roles with the required permissions.

## Benchmarking Instructions

1. Change directory to the benchmarking bin directory:
    ```bash
    cd benchmarking/bin
    ```
1. Edit the parameters in run-all.sh to suite your environment.

1. Execute the benchmarking script:
    ```bash
    ./run-all.sh
    ```
1. Extract the results with a log insights query similar to the following:
    ```
    fields @timestamp as Timestamp, @initDuration as InitDuration, @duration as Duration, Duration as BilledDuration, (@maxMemoryUsed / (1024*1024)) as MaxMemoryUsed
    | parse @log /bmlambda-(?<Runtime>[A-Z][a-z]*)(?<Function>[A-Z][a-z]*)/
    | filter strcontains(@message, "REPORT")
    | sort @timestamp desc
    | limit 6000
    ```    

## Cleanup
 
1. Delete the stack
    ```bash
    make delete
    ```
1. Confirm the stack has been deleted
    ```bash
    aws cloudformation list-stacks --query "StackSummaries[?contains(StackName,'STACK_NAME')].StackStatus"
    ```

----
Copyright 2023 Nigel Slack-Smith. All Rights Reserved.

SPDX-License-Identifier: MIT-0