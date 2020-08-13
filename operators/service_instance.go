package operators

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/credhub-cli/credhub"
)

type BoundApp struct {
	BindingGUID string `json:"binding_guid"`
	AppGUID     string `json:"app_guid"`
	SpaceGUID   string `json:"space_guid"`
}

func (b BoundApp) ActorName() string {
	return fmt.Sprintf("mtls-app:%s", b.AppGUID)
}

type ServicePlan struct {
	Name string `json:"name"`
}

type ServiceInstance struct {
	ServiceInstanceGUID string              `json:"service_instance_guid"`
	ResidentSpaceGUID   string              `json:"resident_space_guid"`
	OrganizationGUID    string              `json:"organization_guid"`
	ServicePlan         ServicePlan         `json:"service_plan"`
	AppBindings         map[string]BoundApp `json:"app_bindings"`
	SecretNames         []string            `json:"secret_names"`
}

func NewServiceInstance(serviceInstanceGUID, residentSpaceGUID, organizationGUID, servicePlanName string) *ServiceInstance {
	return &ServiceInstance{
		ServiceInstanceGUID: serviceInstanceGUID,
		ResidentSpaceGUID:   residentSpaceGUID,
		OrganizationGUID:    organizationGUID,
		ServicePlan: ServicePlan{
			Name: servicePlanName,
		},
	}
}

func (s *ServiceInstance) UpdateSecretsList(secrets map[string]interface{}) {
	s.SecretNames = []string{}
	for secretName, _ := range secrets {
		s.SecretNames = append(s.SecretNames, secretName)
	}
}

func (s *ServiceInstance) DescriptionForUsers() string {
	return fmt.Sprintf("stored secrets: [%s]", strings.Join(s.SecretNames, ", "))
}

func (s *ServiceInstance) Path(prefix string) string {
	return fmt.Sprintf("%s/%s", prefix, s.ServiceInstanceGUID)
}

func serviceInstanceMetaPath(prefix, serviceInstanceGUID string) string {
	return fmt.Sprintf("%s/meta/%s", prefix, serviceInstanceGUID)
}

func (s *ServiceInstance) Save(prefix string, credhubClient *credhub.CredHub) error {
	asCredential, err := s.intoJSONCredential()
	if err != nil {
		return err
	}
	path := serviceInstanceMetaPath(prefix, s.ServiceInstanceGUID)
	_, err = credhubClient.SetJSON(path, asCredential)
	return err
}

func (s *ServiceInstance) Remove(prefix string, credhubClient *credhub.CredHub) error {
	path := serviceInstanceMetaPath(prefix, s.ServiceInstanceGUID)
	return credhubClient.Delete(path)
}

func LoadServiceInstance(serviceInstanceGUID, prefix string, credhubClient *credhub.CredHub) (*ServiceInstance, error) {
	path := serviceInstanceMetaPath(prefix, serviceInstanceGUID)
	credential, err := credhubClient.GetLatestJSON(path)
	if err != nil {
		return nil, err
	}
	return serviceInstanceFromJSONCredential(credential.Value)
}

func (s ServiceInstance) intoJSONCredential() (map[string]interface{}, error) {
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var credential map[string]interface{}
	err = json.Unmarshal(bytes, &credential)
	return credential, err
}

func serviceInstanceFromJSONCredential(credential map[string]interface{}) (*ServiceInstance, error) {
	bytes, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	var serviceInstance ServiceInstance
	err = json.Unmarshal(bytes, &serviceInstance)
	return &serviceInstance, err
}
