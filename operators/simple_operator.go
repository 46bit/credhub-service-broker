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

func (o *SimpleOperator) Provision(ctx context.Context, input provideriface.ProvisionData) (isAsync bool, err error) {
	serviceInstance := NewServiceInstance(
		input.InstanceID,
		input.Details.SpaceGUID,
		input.Details.OrganizationGUID,
	)

	path := serviceInstance.Path(o.prefix)
	if err := ensureNoExistingPermissions(path, o.credhubClient); err != nil {
		return false, err
	}

	var secrets map[string]interface{}
	err = json.Unmarshal(input.Details.RawParameters, &secrets)
	if err != nil {
		return false, err
	}
	if err = writeJSONCredential(path, secrets, o.credhubClient); err != nil {
		return false, err
	}
	serviceInstance.UpdateSecretsList(secrets)

	err = serviceInstance.Save(o.prefix, o.credhubClient)
	return true, err
}

func (o *SimpleOperator) Deprovision(ctx context.Context, input provideriface.DeprovisionData) (isAsync bool, err error) {
	serviceInstance, err := LoadServiceInstance(input.InstanceID, o.prefix, o.credhubClient)
	if err != nil {
		return false, err
	}

	path := serviceInstance.Path(o.prefix)
	if err := deletePermissions(path, o.credhubClient); err != nil {
		return false, err
	}

	if err := deleteCredential(path, o.credhubClient); err != nil {
		return false, err
	}

	err = serviceInstance.Remove(o.prefix, o.credhubClient)
	return false, err
}

func (o *SimpleOperator) Update(ctx context.Context, input provideriface.UpdateData) (isAsync bool, err error) {
	serviceInstance, err := LoadServiceInstance(input.InstanceID, o.prefix, o.credhubClient)
	if err != nil {
		return false, err
	}

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

	err = serviceInstance.Save(o.prefix, o.credhubClient)
	return true, err
}

func (o *SimpleOperator) Bind(ctx context.Context, input provideriface.BindData) (binding brokerapi.Binding, err error) {
	serviceInstance, err := LoadServiceInstance(input.InstanceID, o.prefix, o.credhubClient)
	if err != nil {
		return brokerapi.Binding{}, err
	}

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
	err = serviceInstance.Save(o.prefix, o.credhubClient)
	return brokerapi.Binding{
		Credentials: map[string]string{
			"credhub-ref": path,
		},
	}, err
}

func (o *SimpleOperator) Unbind(ctx context.Context, input provideriface.UnbindData) (unbinding brokerapi.UnbindSpec, err error) {
	serviceInstance, err := LoadServiceInstance(input.InstanceID, o.prefix, o.credhubClient)
	if err != nil {
		return brokerapi.UnbindSpec{}, err
	}
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
	err = serviceInstance.Save(o.prefix, o.credhubClient)
	return brokerapi.UnbindSpec{}, nil
}

func (o *SimpleOperator) LastOperation(ctx context.Context, input provideriface.LastOperationData) (state brokerapi.LastOperationState, description string, err error) {
	serviceInstance, err := LoadServiceInstance(input.InstanceID, o.prefix, o.credhubClient)
	if err != nil {
		return brokerapi.Failed, "", err
	}
	return brokerapi.Succeeded, serviceInstance.DescriptionForUsers(), nil
}
