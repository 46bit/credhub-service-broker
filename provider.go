package main

import (
	"context"
	"encoding/json"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/lager"
	provideriface "github.com/alphagov/paas-service-broker-base/provider"
	"github.com/pivotal-cf/brokerapi"
)

type CredHubProvider struct {
	prefix        string
	credhubClient *credhub.CredHub
	logger        lager.Logger
}

func NewCredHubProvider(prefix string, credhubClient *credhub.CredHub, logger lager.Logger) *CredHubProvider {
	logger = logger.Session("credhub-provider", lager.Data{"prefix": prefix})
	return &CredHubProvider{prefix, credhubClient, logger}
}

func (p *CredHubProvider) Provision(ctx context.Context, provisionData provideriface.ProvisionData) (dashboardURL, operationData string, isAsync bool, err error) {
	serviceInstance := NewServiceInstance(
		provisionData.InstanceID,
		provisionData.Details.SpaceGUID,
		provisionData.Details.OrganizationGUID,
	)

	path := serviceInstance.Path(p.prefix)
	if err := ensureNoExistingPermissions(path, p.credhubClient); err != nil {
		return "", "", false, err
	}

	var secrets map[string]interface{}
	err = json.Unmarshal(provisionData.Details.RawParameters, &secrets)
	if err != nil {
		return "", "", false, err
	}
	if err = writeJSONCredential(path, secrets, p.credhubClient); err != nil {
		return "", "", false, err
	}
	serviceInstance.UpdateSecretsList(secrets)

	err = serviceInstance.Save(p.prefix, p.credhubClient)
	return "", "", true, err
}

func (p *CredHubProvider) Deprovision(ctx context.Context, deprovisionData provideriface.DeprovisionData) (operationData string, isAsync bool, err error) {
	serviceInstance, err := LoadServiceInstance(deprovisionData.InstanceID, p.prefix, p.credhubClient)
	if err != nil {
		return "", false, err
	}

	path := serviceInstance.Path(p.prefix)
	if err := deletePermissions(path, p.credhubClient); err != nil {
		return "", false, err
	}

	if err := deleteCredential(path, p.credhubClient); err != nil {
		return "", false, err
	}

	err = serviceInstance.Remove(p.prefix, p.credhubClient)
	return "", false, err
}

func (p *CredHubProvider) Update(ctx context.Context, updateData provideriface.UpdateData) (operationData string, isAsync bool, err error) {
	serviceInstance, err := LoadServiceInstance(updateData.InstanceID, p.prefix, p.credhubClient)
	if err != nil {
		return "", false, err
	}

	var secrets map[string]interface{}
	err = json.Unmarshal(updateData.Details.RawParameters, &secrets)
	if err != nil {
		return "", false, err
	}
	path := serviceInstance.Path(p.prefix)
	if err = writeJSONCredential(path, secrets, p.credhubClient); err != nil {
		return "", false, err
	}
	serviceInstance.UpdateSecretsList(secrets)

	err = serviceInstance.Save(p.prefix, p.credhubClient)
	return "", true, err
}

func (p *CredHubProvider) Bind(ctx context.Context, bindData provideriface.BindData) (binding brokerapi.Binding, err error) {
	serviceInstance, err := LoadServiceInstance(bindData.InstanceID, p.prefix, p.credhubClient)
	if err != nil {
		return brokerapi.Binding{}, err
	}

	appBinding := BoundApp{
		BindingGUID: bindData.BindingID,
		AppGUID:     bindData.Details.AppGUID,
		SpaceGUID:   bindData.Details.BindResource.SpaceGuid,
	}
	actor := appBinding.ActorName()
	path := serviceInstance.Path(p.prefix)
	if err = grantReadAccess(actor, path, p.credhubClient); err != nil {
		return brokerapi.Binding{}, err
	}

	if serviceInstance.AppBindings == nil {
		serviceInstance.AppBindings = map[string]BoundApp{}
	}
	serviceInstance.AppBindings[appBinding.BindingGUID] = appBinding
	err = serviceInstance.Save(p.prefix, p.credhubClient)
	return brokerapi.Binding{
		Credentials: map[string]string{
			"credhub-ref": path,
		},
	}, err
}

func (p *CredHubProvider) Unbind(ctx context.Context, unbindData provideriface.UnbindData) (unbinding brokerapi.UnbindSpec, err error) {
	serviceInstance, err := LoadServiceInstance(unbindData.InstanceID, p.prefix, p.credhubClient)
	if err != nil {
		return brokerapi.UnbindSpec{}, err
	}
	appBinding, ok := serviceInstance.AppBindings[unbindData.BindingID]
	if !ok {
		return brokerapi.UnbindSpec{}, nil
	}

	actor := appBinding.ActorName()
	path := serviceInstance.Path(p.prefix)
	if err = revokeReadAccess(actor, path, p.credhubClient); err != nil {
		return brokerapi.UnbindSpec{}, err
	}

	delete(serviceInstance.AppBindings, appBinding.BindingGUID)
	err = serviceInstance.Save(p.prefix, p.credhubClient)
	return brokerapi.UnbindSpec{}, nil
}

func (p *CredHubProvider) LastOperation(ctx context.Context, lastOperationData provideriface.LastOperationData) (state brokerapi.LastOperationState, description string, err error) {
	serviceInstance, err := LoadServiceInstance(lastOperationData.InstanceID, p.prefix, p.credhubClient)
	if err != nil {
		return brokerapi.Failed, "", err
	}
	return brokerapi.Succeeded, serviceInstance.DescriptionForUsers(), nil
}
