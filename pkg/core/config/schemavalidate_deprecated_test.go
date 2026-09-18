package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/jsonschema"
)

func TestSchemaDeprecationSupport(t *testing.T) {
	type fixture struct {
		SettingName           string `yaml:"settingname,omitempty"`
		DeprecatedSettingName string `yaml:"settingName,omitempty" jsonschema:"-"`
	}

	for _, tt := range []struct {
		name     string
		strict   bool
		disallow bool
		key      string
		severity Severity
		message  string
	}{
		{"supported deprecation", false, false, "settingName", SeverityWarning, `"settingName" is deprecated in favor of "settingname"`},
		{"strict deprecation", true, false, "settingName", SeverityError, `"settingName" is deprecated in favor of "settingname"`},
		{"disabled deprecation support", false, true, "settingName", SeverityError, `unknown key "settingName"`},
		{"removed v0 key", false, false, "scmID", SeverityError, `unknown key "scmID"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := newSchemaRegistry()
			require.NoError(t, err)
			registry.root, err = jsonschema.Compile(schemaBaseID+"/test-deprecation", specReflector().Reflect(fixture{}))
			require.NoError(t, err)
			registry.deprecated[nodeRoot] = deprecatedFieldKeys([]reflect.Type{reflect.TypeOf(fixture{})})

			options := DefaultSchemaValidationOptions()
			options.Strict = tt.strict
			if tt.disallow {
				options.AllowDeprecated = false
			}
			validation := schemaValidator{registry: registry, options: options}
			validation.validateDocument(map[string]interface{}{tt.key: "value"})
			require.Len(t, validation.problems, 1)
			require.Equal(t, tt.severity, validation.problems[0].Severity)
			require.Contains(t, validation.problems[0].Message, tt.message)
		})
	}
}
