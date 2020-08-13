package main

import (
	"context"
	"fmt"

	"github.com/46bit/credhub-service-broker/operators"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/lager"
	provideriface "github.com/alphagov/paas-service-broker-base/provider"
	"github.com/pivotal-cf/brokerapi"
)

type BrokerProvider struct {
	operators     map[string]operators.Operator
	prefix        string
	credhubClient *credhub.CredHub
	logger        lager.Logger
}

func NewBrokerProvider(operators map[string]operators.Operator, prefix string, credhubClient *credhub.CredHub, logger lager.Logger) *BrokerProvider {
	logger = logger.Session("broker-provider")
	return &BrokerProvider{operators, prefix, credhubClient, logger}
}

func (p *BrokerProvider) Provision(ctx context.Context, input provideriface.ProvisionData) (dashboardURL, operationData string, isAsync bool, err error) {
	operator, ok := p.operators[input.Plan.Name]
	if !ok {
		return "", "", false, fmt.Errorf("no operator found for service plan '%s'", input.Plan.Name)
	}

	serviceInstance, isAsync, err := operator.Provision(ctx, input.Plan.Name, input)
	if err != nil {
		return "", "", false, err
	}
	err = serviceInstance.Save(p.prefix, p.credhubClient)
	return "", "", isAsync, err
}

func (p *BrokerProvider) Deprovision(ctx context.Context, input provideriface.DeprovisionData) (operationData string, isAsync bool, err error) {
	serviceInstance, operator, err := p.loadServiceInstance(input.InstanceID)
	if err != nil {
		return "", false, err
	}

	isAsync, err = (*operator).Deprovision(ctx, serviceInstance, input)
	if err != nil {
		return "", false, err
	}
	err = serviceInstance.Remove(p.prefix, p.credhubClient)
	if err != nil {
		return "", false, err
	}
	return "", isAsync, nil
}

func (p *BrokerProvider) Update(ctx context.Context, input provideriface.UpdateData) (operationData string, isAsync bool, err error) {
	serviceInstance, operator, err := p.loadServiceInstance(input.InstanceID)
	if err != nil {
		return "", false, err
	}

	isAsync, err = (*operator).Update(ctx, serviceInstance, input)
	if err != nil {
		return "", false, err
	}
	err = serviceInstance.Save(p.prefix, p.credhubClient)
	return "", isAsync, err
}

func (p *BrokerProvider) Bind(ctx context.Context, input provideriface.BindData) (binding brokerapi.Binding, err error) {
	serviceInstance, operator, err := p.loadServiceInstance(input.InstanceID)
	if err != nil {
		return brokerapi.Binding{}, err
	}

	binding, err = (*operator).Bind(ctx, serviceInstance, input)
	if err != nil {
		return brokerapi.Binding{}, err
	}
	err = serviceInstance.Save(p.prefix, p.credhubClient)
	if err != nil {
		return brokerapi.Binding{}, err
	}
	return binding, nil
}

func (p *BrokerProvider) Unbind(ctx context.Context, input provideriface.UnbindData) (unbinding brokerapi.UnbindSpec, err error) {
	serviceInstance, operator, err := p.loadServiceInstance(input.InstanceID)
	if err != nil {
		return brokerapi.UnbindSpec{}, err
	}

	unbinding, err = (*operator).Unbind(ctx, serviceInstance, input)
	if err != nil {
		return brokerapi.UnbindSpec{}, err
	}
	err = serviceInstance.Save(p.prefix, p.credhubClient)
	if err != nil {
		return brokerapi.UnbindSpec{}, err
	}
	return unbinding, nil
}

func (p *BrokerProvider) LastOperation(ctx context.Context, input provideriface.LastOperationData) (state brokerapi.LastOperationState, description string, err error) {
	serviceInstance, operator, err := p.loadServiceInstance(input.InstanceID)
	if err != nil {
		return brokerapi.Failed, "", err
	}

	// FIXME: Rename LastOperation on the operator to make clear it's just to control the
	// text in `cf service`?
	return (*operator).LastOperation(ctx, serviceInstance, input)
}

func (p *BrokerProvider) loadServiceInstance(instanceGUID string) (serviceInstance *operators.ServiceInstance, operator *operators.Operator, err error) {
	serviceInstance, err = operators.LoadServiceInstance(instanceGUID, p.prefix, p.credhubClient)
	if err != nil {
		return nil, nil, err
	}
	operatorValue, ok := p.operators[serviceInstance.ServicePlan.Name]
	if !ok {
		return nil, nil, fmt.Errorf("no operator found for service plan '%s'", serviceInstance.ServicePlan.Name)
	}
	return serviceInstance, &operatorValue, nil
}
