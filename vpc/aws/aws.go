package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type VPC struct {
	client aws.Config
	ec2    *ec2.Client
	ctx    context.Context
}

func Create(ctx context.Context, cidr string) (*VPC, error) {
	vpc := &VPC{}

	defaultConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	vpc.client = defaultConfig
	vpc.ec2 = ec2.NewFromConfig(defaultConfig)
	vpc.ctx = ctx
	return vpc, nil
}

func (v *VPC) createVPC(cidr string) (string, error) {
	out, err := v.ec2.CreateVpc(v.ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String(cidr),
	})
	if err != nil {
		return "", err
	}
	if _, err = v.ec2.ModifyVpcAttribute(v.ctx, &ec2.ModifyVpcAttributeInput{
		VpcId:              out.Vpc.VpcId,
		EnableDnsHostnames: &types.AttributeBooleanValue{Value: aws.Bool(true)},
		EnableDnsSupport:   &types.AttributeBooleanValue{Value: aws.Bool(true)},
	}); err != nil {
		return "", err
	}

	return *out.Vpc.VpcId, nil
}

func (v *VPC) destroyVPC(id string) error {
	_, err := v.ec2.DeleteVpc(v.ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(id),
	})
	return err
}

func (v *VPC) createNat() {

}

func (v *VPC) destroyNat() {

}

func (v *VPC) createSubnet(cidr string, vpcID string) error {
	out, err := v.ec2.CreateSubnet(v.ctx, &ec2.CreateSubnetInput{
		CidrBlock:        aws.String(cidr),
		VpcId:            aws.String(vpcID),
		AvailabilityZone: aws.String("us-east-1a"),
	})
	if err != nil {
		return err
	}
	_, err = v.ec2.CreateTags(v.ctx, &ec2.CreateTagsInput{
		Resources: []string{*out.Subnet.SubnetId},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test"),
			},
		},
	})
	return err
}

func (v *VPC) destroySubnet(id string) error {
	if _, err := v.ec2.DeleteSubnet(v.ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(id),
	}); err != nil {
		return err
	}
	return nil
}

func (v *VPC) createVPCPeer() {

}

func (v *VPC) disconnectVPCPeer() {

}

func (v *VPC) createFirewall() {
}

func (v *VPC) destroyFirewall() {

}

type FirewallRule struct {
}

func (v *VPC) createFirewallRule() {

}

func (v *VPC) destroyFirewallRule() {

}
