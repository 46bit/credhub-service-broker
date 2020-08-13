package operators

import (
	"context"

	provideriface "github.com/alphagov/paas-service-broker-base/provider"
	"github.com/pivotal-cf/brokerapi"
)

type Operator interface {
	Provision(context.Context, provideriface.ProvisionData) (isAsync bool, err error)
	Deprovision(context.Context, provideriface.DeprovisionData) (isAsync bool, err error)
	Update(context.Context, provideriface.UpdateData) (isAsync bool, err error)
	Bind(context.Context, provideriface.BindData) (brokerapi.Binding, error)
	Unbind(context.Context, provideriface.UnbindData) (brokerapi.UnbindSpec, error)
	LastOperation(context.Context, provideriface.LastOperationData) (state brokerapi.LastOperationState, description string, err error)
}
