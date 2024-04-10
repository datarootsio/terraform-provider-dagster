package types

import (
	"fmt"
	"strings"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
)

var grantMap = map[string]schema.PermissionGrant{
	"VIEWER":   schema.PermissionGrantViewer,
	"LAUNCHER": schema.PermissionGrantLauncher,
	"EDITOR":   schema.PermissionGrantEditor,
	"ADMIN":    schema.PermissionGrantAdmin,
}

func ConvertToGrantEnum(grant string) (schema.PermissionGrant, error) {
	enum, ok := grantMap[strings.ToUpper(grant)]
	if !ok {
		return "", fmt.Errorf("could not convert (%s) to grant enum", grant)
	}

	return enum, nil
}

func DeploymentGrantEnumValues() []string {
	return []string{"VIEWER", "LAUNCHER", "EDITOR", "ADMIN"}
}

func LocationGrantEnumValues() []string {
	return []string{"LAUNCHER", "EDITOR", "ADMIN"}
}
