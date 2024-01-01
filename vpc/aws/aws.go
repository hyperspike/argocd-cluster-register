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

func Create(ctx context.Context) (*VPC, error) {
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

func (v *VPC) destroyVPC() {

}

func (v *VPC) createNat() {

}

func (v *VPC) destroyNat() {

}

func (v *VPC) createSubnet() {

}

func (v *VPC) destroySubnet() {

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
