# Cloud Resource Discovery

Potential Red Canary customers require an estimate of how much their Red Canary services will be.

This application enables customers to scan their AWS accounts and provides a count of specific objects which are relevant to the billing process.

This application is ran from within the customer's CloudShell env. 

## System Requirements

 - AWS credentials with read-only access to the account or organization you wish to scan. A user with the `SecurityAudit` managed policy is sufficient, but customers may wish to create a custom policy with more restrictive permissions.

<details><summary>Suggested policy with minimal permissions (click to expand):</summary>

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Resource": "*",
            "Action": [
                "organizations:ListAccounts",
                "ec2:DescribeRegions",
                "s3:ListBucket",
                "s3:GetBucketLocation",
                "s3:GetBucketNotification",
                "s3:ListAllMyBuckets",
                "dynamodb:ListTables",
                "dynamodb:ListGlobalTables",
                "ec2:DescribeVolumes",
                "ec2:DescribeInstances",
                "ecr:DescribeRepositories",
                "ecr:ListImages",
                "ecr-public:DescribeRepositories",
                "ecr-public:DescribeImages",
                "ecs:ListClusters",
                "ecs:ListServices",
                "ecs:DescribeServices",
                "ecs:DescribeTaskDefinition",
                "ecs:DescribeClusters",
                "elasticfilesystem:DescribeFileSystems",
                "eks:ListClusters",
                "lambda:ListFunctions",
                "rds:DescribeDBInstances",
                "cloudformation:ListResources",
                "cloudformation:DescribeStacks",
                "cloudtrail:DescribeTrails",
            ]
        }
    ]
}
```

</details>

## Setup

```bash
git clone https://github.com/redcanaryco/red-canary-aws-resource-discovery.git
cd red-canary-aws-resource-discovery
go mod tidy
```

After, navigate into the cmd/ folder and build. To compile the binary for use with cloudshell, please run the following from within cmd/
```bash
GOOS=linux GOARCH=amd64 go build -o enumerate-resources
```
In the above case, I called my binary "enumerate-resources".

Navigate into the AWS env with CloudShell and the discovery role setup. 

Select the region closest to you from the dropdown, click Actions > Upload File and select the built binary. By default the binary is build in the cmd/ directory. 

After upload, run the sha256sum command from both the cmd/ and cloudshell env after upload. Verify the SHA is the same.
```bash 
$ sha256sum enumerate-resources
```

## Usage

We only provide official support for running this application from cloudshell using either the provided binary or having the code as is compiled into a binary:


If you would like to scan the whole organization, run the binary as is:

```bash
./enumerate-resources
```

If you would like to scan the whole organization and display the cloudtrail information for the accounts, run the binary with the AWS_TRAIL flag set to true. By default, it's set to false:

```bash
./enumerate-resources --AWS_TRAIL="true"
```

If you would like to scan the whole organization and exclude certain accounts, run the binary with the EXCLUDE flag with one or many accounts separated by a comma.

```bash
./enumerate-resources --AWS_TRAIL="true" --EXCLUDE="123456789"
```

OR

```bash
./enumerate-resources --AWS_TRAIL="true" --EXCLUDE="123456789,423456789,523456789"
```


If you would like to scan a different account, you can specify the profile name via the `AWS_ACCOUNT_ID` environment variable.

```bash
./enumerate-resources 
    --AWS_ACCOUNT_ID="123456789" 
```

Provided below are the different flags that may be set for any given run. The flags work in any combination with each other -- other than selecting a single account ID (AWS_ACCOUNT_ID) and excluding (EXCLUDE). 

```bash
./enumerate-resources 
    --AWS_ROLE_ARN="arn:aws:iam::123456789:role/red-canary-resource-discovery-role" 
    --AWS_ACCOUNT_ID="123456789" 
    --AWS_REGION="us-east-1" 
    --AWS_ROLE_NAME="red-canary-resource-discovery-role"
    --AWS_TRAIL="true"
    --EXCLUDE="123456789,5236756789,2344675689,3446756890"
```

The application will display summarized output in the console, and produce a CSV report in the current working directory.

```bash
$ Red Canary - AWS Resource Discovery Scan Progress: 34 / 34

Scanned 2 AWS accounts.

ResourceType                Count   
Storage Buckets             10      
Container Hosts             35       
Databases                   9       
Non-OS Disks                12      
Serverless Containers       25       
Serverless Functions        16       
Virtual Machines            65      
Container Registry Images   13      

Scan completed in 45 seconds.

$ ls
aws-resource-discovery.csv

$ cat aws-resource-discovery.csv

...
123456789,us-east-1,AWS::S3::Bucket,3
123456789,us-east-1,AWS::RDS::DBInstance,0
123456789,us-east-1,AWS::ECS::Cluster,1
123456789,us-east-1,AWS::EKS::Cluster,2
123456789,us-east-1,AWS::ECR::PublicRepository,1
123456789,us-east-1,AWS::EC2::Instance,3
123456789,us-east-1,AWS::ECR::Repository,0
123456789,us-east-1,AWS::EFS::FileSystem,0
123456789,us-east-1,AWS::DynamoDB::Table,0
123456789,us-east-1,AWS::Lambda::Function,0
123456789,us-east-1,AWS::EC2::Volume,3
...
```

## Troubleshooting

### Error: `The security token included in the request is invalid.`

This error indicates that the AWS credentials you provided are invalid. Please double-check that you have provided the correct credentials.

### Error: `You are not authorized to perform this operation.`

This error indicates that the AWS credentials you provided do not have sufficient permissions to perform the requested operation. Typically this kind of opaque error is caused by missing EC2 permissions. Please ensure that your credentials have the appropriate EC2 permissions as described above.


## License

Copyright (c) 2024 Red Canary, Inc. All rights reserved.
