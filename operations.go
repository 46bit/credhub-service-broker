package main

import (
	"fmt"

	"code.cloudfoundry.org/credhub-cli/credhub"
)

func ensureNoExistingPermissions(path string, credhub *credhub.CredHub) error {
	permissions, err := credhub.GetPermissions(path)
	if err == nil && len(permissions) > 0 {
		return fmt.Errorf("path '%s' had existing permissions", path)
	}
	return nil
}

func writeJSONCredential(path string, json map[string]interface{}, credhub *credhub.CredHub) error {
	_, err := credhub.SetJSON(path, json)
	return err
}

func ensureCredentialExists(path string, credhub *credhub.CredHub) error {
	_, err := credhub.GetLatestVersion(fmt.Sprintf("%s", path))
	return err
}

func deletePermissions(path string, credhub *credhub.CredHub) error {
	permissions, err := credhub.GetPermissions(path)
	if err != nil {
		return err
	}
	for _, permission := range permissions {
		// FIXME: This is awful
		permissionObjectWithUUID, err := credhub.GetPermissionByPathActor(path, permission.Actor)
		if err != nil {
			return err
		}
		if _, err = credhub.DeletePermission(permissionObjectWithUUID.UUID); err != nil {
			return err
		}
	}
	return nil
}

func deleteCredential(path string, credhub *credhub.CredHub) error {
	return credhub.Delete(path)
}

func grantReadAccess(actor, path string, credhub *credhub.CredHub) error {
	_, err := credhub.AddPermission(path, actor, []string{"read"})
	return err
}

func revokeReadAccess(actor, path string, credhub *credhub.CredHub) error {
	existingPermission, err := credhub.GetPermissionByPathActor(path, actor)
	if err != nil {
		return err
	}
	_, err = credhub.DeletePermission(existingPermission.UUID)
	return err
}
