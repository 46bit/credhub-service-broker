package operators

import (
	"context"
	"encoding/json"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/lager"
	provideriface "github.com/alphagov/paas-service-broker-base/provider"
	"github.com/pivotal-cf/brokerapi"
)

type SimpleOperator struct {
	prefix        string
	credhubClient *credhub.CredHub
	logger        lager.Logger
}

func NewSimpleOperator(prefix string, credhubClient *credhub.CredHub, logger lager.Logger) *SimpleOperator {
	logger = logger.Session("simple-operator", lager.Data{"prefix": prefix})
	return &SimpleOperator{prefix, credhubClient, logger}
}

func (o *SimpleOperator) Provision(ctx context.Context, servicePlanName string, input provideriface.ProvisionData) (serviceInstance *ServiceInstance, isAsync bool, err error) {
	serviceInstance = NewServiceInstance(
		input.InstanceID,
		input.Details.SpaceGUID,
		input.Details.OrganizationGUID,
		servicePlanName,
	)

	path := serviceInstance.Path(o.prefix)
	if err := ensureNoExistingPermissions(path, o.credhubClient); err != nil {
		return nil, false, err
	}

	var secrets map[string]interface{}
	err = json.Unmarshal(input.Details.RawParameters, &secrets)
	if err != nil {
		return nil, false, err
	}
	if err = writeJSONCredential(path, secrets, o.credhubClient); err != nil {
		return nil, false, err
	}
	serviceInstance.UpdateSecretsList(secrets)
	return serviceInstance, true, nil
}

func (o *SimpleOperator) Deprovision(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.DeprovisionData) (isAsync bool, err error) {
	path := serviceInstance.Path(o.prefix)
	if err := deletePermissions(path, o.credhubClient); err != nil {
		return false, err
	}

	if err := deleteCredential(path, o.credhubClient); err != nil {
		return false, err
	}
	return false, nil
}

func (o *SimpleOperator) Update(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.UpdateData) (isAsync bool, err error) {
	var secrets map[string]interface{}
	err = json.Unmarshal(input.Details.RawParameters, &secrets)
	if err != nil {
		return false, err
	}
	path := serviceInstance.Path(o.prefix)
	if err = writeJSONCredential(path, secrets, o.credhubClient); err != nil {
		return false, err
	}
	serviceInstance.UpdateSecretsList(secrets)
	return true, nil
}

func (o *SimpleOperator) Bind(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.BindData) (binding brokerapi.Binding, err error) {
	appBinding := BoundApp{
		BindingGUID: input.BindingID,
		AppGUID:     input.Details.AppGUID,
		SpaceGUID:   input.Details.BindResource.SpaceGuid,
	}
	actor := appBinding.ActorName()
	path := serviceInstance.Path(o.prefix)
	if err = grantReadAccess(actor, path, o.credhubClient); err != nil {
		return brokerapi.Binding{}, err
	}

	if serviceInstance.AppBindings == nil {
		serviceInstance.AppBindings = map[string]BoundApp{}
	}
	serviceInstance.AppBindings[appBinding.BindingGUID] = appBinding
	return brokerapi.Binding{
		Credentials: map[string]string{
			"credhub-ref": path,
		},
	}, nil
}

func (o *SimpleOperator) Unbind(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.UnbindData) (unbinding brokerapi.UnbindSpec, err error) {
	appBinding, ok := serviceInstance.AppBindings[input.BindingID]
	if !ok {
		return brokerapi.UnbindSpec{}, nil
	}

	actor := appBinding.ActorName()
	path := serviceInstance.Path(o.prefix)
	if err = revokeReadAccess(actor, path, o.credhubClient); err != nil {
		return brokerapi.UnbindSpec{}, err
	}

	delete(serviceInstance.AppBindings, appBinding.BindingGUID)
	return brokerapi.UnbindSpec{}, nil
}

func (o *SimpleOperator) LastOperation(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.LastOperationData) (state brokerapi.LastOperationState, description string, err error) {
	return brokerapi.Succeeded, serviceInstance.DescriptionForUsers(), nil
}
