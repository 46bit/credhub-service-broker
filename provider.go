package main

import (
	"context"

	"github.com/46bit/credhub-service-broker/operators"

	"code.cloudfoundry.org/lager"
	provideriface "github.com/alphagov/paas-service-broker-base/provider"
	"github.com/pivotal-cf/brokerapi"
)

type BrokerProvider struct {
	operator operators.Operator
	logger   lager.Logger
}

func NewBrokerProvider(operator operators.Operator, logger lager.Logger) *BrokerProvider {
	logger = logger.Session("broker-provider")
	return &BrokerProvider{operator, logger}
}

func (p *BrokerProvider) Provision(ctx context.Context, input provideriface.ProvisionData) (dashboardURL, operationData string, isAsync bool, err error) {
	isAsync, err = p.operator.Provision(ctx, input)
	return "", "", isAsync, err
}

func (p *BrokerProvider) Deprovision(ctx context.Context, input provideriface.DeprovisionData) (operationData string, isAsync bool, err error) {
	isAsync, err = p.operator.Deprovision(ctx, input)
	return "", isAsync, err
}

func (p *BrokerProvider) Update(ctx context.Context, input provideriface.UpdateData) (operationData string, isAsync bool, err error) {
	isAsync, err = p.operator.Update(ctx, input)
	return "", isAsync, err
}

func (p *BrokerProvider) Bind(ctx context.Context, input provideriface.BindData) (binding brokerapi.Binding, err error) {
	return p.operator.Bind(ctx, input)
}

func (p *BrokerProvider) Unbind(ctx context.Context, input provideriface.UnbindData) (unbinding brokerapi.UnbindSpec, err error) {
	return p.operator.Unbind(ctx, input)
}

func (p *BrokerProvider) LastOperation(ctx context.Context, input provideriface.LastOperationData) (state brokerapi.LastOperationState, description string, err error) {
	return p.operator.LastOperation(ctx, input)
}
