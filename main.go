package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/ec2/types"
    "github.com/joho/godotenv"
)

func main() {
    // Load the .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    // Retrieve environment variables
    awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
    awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

    // Load the shared AWS configuration
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
        config.WithCredentialsProvider(
            aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
                return aws.Credentials{
                    AccessKeyID:     awsAccessKeyID,
                    SecretAccessKey: awsSecretAccessKey,
                }, nil
            }),
        ),
    )
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }

    // Create an EC2 service client
    svc := ec2.NewFromConfig(cfg)

    securitygroupID, err := CreateSecurityGroup(svc, "sg-s3-scan", "Security group for s3-scan instance")
    if err != nil {
        fmt.Println("Failed to create security group: ", err)
        return
    }
    fmt.Println("Created security group: ",securitygroupID)

    // Define the user data script to create the .env file, download the script and requirements file, install dependencies, and run the Python script
    userDataScript := fmt.Sprintf(`#!/bin/bash
    sudo apt update -y
    sudo apt install -y python3
    echo "export AWS_ACCESS_KEY_ID=%s" >> /tmp/.env
    echo "export AWS_SECRET_ACCESS_KEY=%s" >> /tmp/.env
    curl -o /tmp/startup-script.py https://raw.githubusercontent.com/ankan-0610/pii-scan-script/master/extract-pii.py
    curl -o /tmp/requirements.txt https://raw.githubusercontent.com/ankan-0610/pii-scan-script/master/requirements.txt
    python3 -m pip install -r /tmp/requirements.txt
    source /tmp/.env
    python3 /tmp/startup-script.py`, awsAccessKeyID, awsSecretAccessKey)

    // Run the instance
    runResult, err := svc.RunInstances(context.TODO(), &ec2.RunInstancesInput{
        ImageId:      aws.String("ami-0c02fb55956c7d316"), // Amazon Linux 2 AMI
        InstanceType: types.InstanceTypeT2Micro,
        MinCount:     aws.Int32(1),
        MaxCount:     aws.Int32(1),
        UserData:     aws.String(userDataScript),
        // KeyName:      aws.String("your-key-pair-name"), // Replace with your key pair name
        SecurityGroupIds: []string{
            securitygroupID, // Replace with your security group id
        },
    })

    if err != nil {
        fmt.Println("Could not create instance", err)
        return
    }

    fmt.Println("Created instance", *runResult.Instances[0].InstanceId)
}
