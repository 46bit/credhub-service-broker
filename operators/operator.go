package operators

import (
	"context"

	provideriface "github.com/alphagov/paas-service-broker-base/provider"
	"github.com/pivotal-cf/brokerapi"
)

type Operator interface {
	Provision(ctx context.Context, servicePlanName string, input provideriface.ProvisionData) (serviceInstance *ServiceInstance, isAsync bool, err error)
	Deprovision(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.DeprovisionData) (isAsync bool, err error)
	Update(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.UpdateData) (isAsync bool, err error)
	Bind(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.BindData) (brokerapi.Binding, error)
	Unbind(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.UnbindData) (brokerapi.UnbindSpec, error)
	LastOperation(ctx context.Context, serviceInstance *ServiceInstance, input provideriface.LastOperationData) (state brokerapi.LastOperationState, description string, err error)
}
