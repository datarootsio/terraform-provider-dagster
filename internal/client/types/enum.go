package types

import (
	"fmt"
	"strings"

	"github.com/datarootsio/terraform-provider-dagster/internal/client/schema"
	"golang.org/x/exp/maps"
)

var grantMap = map[string]schema.PermissionGrant{
	"VIEWER":   schema.PermissionGrantViewer,
	"LAUNCHER": schema.PermissionGrantLauncher,
	"EDITOR":   schema.PermissionGrantEditor,
	"ADMIN":    schema.PermissionGrantAdmin,
	"AGENT":    schema.PermissionGrantAgent,
}

func ConvertToGrantEnum(grant string) (schema.PermissionGrant, error) {
	enum, ok := grantMap[strings.ToUpper(grant)]
	if !ok {
		return "", fmt.Errorf("could not convert (%s) to grant enum", grant)
	}

	return enum, nil
}

func GrantEnumValues() []string {
	return maps.Keys(grantMap)
}
