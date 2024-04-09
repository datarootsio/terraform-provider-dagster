package types

import (
	"fmt"
	"strings"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
)

func ConvertToGrantEnum(grant string) (schema.PermissionGrant, error) {
	grantMap := map[string]schema.PermissionGrant{
		"VIEWER":   schema.PermissionGrantViewer,
		"LAUNCHER": schema.PermissionGrantLauncher,
		"EDITOR":   schema.PermissionGrantEditor,
		"ADMIN":    schema.PermissionGrantAdmin,
		"AGENT":    schema.PermissionGrantAgent,
	}

	enum, ok := grantMap[strings.ToUpper(grant)]
	if !ok {
		return "", fmt.Errorf("could not convert (%s) to grant enum", grant)
	}

	return enum, nil
}
