package placeholders_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	api_v1 "github.com/vision-cli/api/v1"
	v1 "github.com/vision-cli/api/v1"
	"github.com/vision-cli/vision-plugin-graphql-v1/placeholders"
)

func TestSetupPlaceholders_WithValidName_ReturnsCorrectPlaceholders(t *testing.T) {
	r := api_v1.PluginRequest{
		Args: []string{"create", "mything"},
	}
	result, err := placeholders.SetupPlaceholders(r)
	require.NoError(t, err)
	expected := &v1.PluginPlaceholders{
		ProjectRoot:               "mything",
		ProjectName:               "mything",
		ProjectDirectory:          "mything",
		ProjectFqn:                "mything",
		Registry:                  "",
		Remote:                    "",
		Branch:                    "",
		Version:                   "",
		ServicesFqn:               "",
		ServicesDirectory:         "",
		GatewayServiceName:        "",
		GatewayFqn:                "",
		GraphqlServiceName:        "",
		GraphqlFqn:                "",
		LibsFqn:                   "mything/libs",
		LibsDirectory:             "",
		ServiceNamespace:          "",
		ServiceVersionedNamespace: "",
		ServiceName:               "",
		ServiceFqn:                "",
		ServiceDirectory:          "",
		InfraDirectory:            "",
		ProtoPackage:              "",
		Deployment:                "",
	}
	assert.Equal(t, expected, result)
}
