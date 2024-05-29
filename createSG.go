package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
)

func CreateSecurityGroup(svc *ec2.Client, securityGroupName string, description string) (string, error) {
	securityGroupInput := &ec2.CreateSecurityGroupInput{
		Description: aws.String(description),
		GroupName:   aws.String(securityGroupName),
	}

	// Create security group
	securityGroupResult, err := svc.CreateSecurityGroup(context.Background(), securityGroupInput)
	if err != nil {
		return "", err
	}

	result, err := svc.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    securityGroupResult.GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(80), // HTTP port
		ToPort:     aws.Int32(80),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	if err != nil {
		// Handle error
		fmt.Println("Failed to authorize security group ingress: ", err)
	}

	println(result)

	result, err = svc.AuthorizeSecurityGroupIngress(context.Background(), &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    securityGroupResult.GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(443), // HTTPS port
		ToPort:     aws.Int32(443),
		CidrIp:     aws.String("0.0.0.0/0"),
	})

	if err != nil {
		// Handle error
		fmt.Println("Failed to authorize security group ingress: ", err)
	}

	println(result)

	return *securityGroupResult.GroupId, nil
}