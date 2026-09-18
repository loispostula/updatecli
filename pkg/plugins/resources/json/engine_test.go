package json

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEngineSelection(t *testing.T) {
	for _, engine := range []string{"default", "", "dasel", "dasel/v3", "dasel/v1", "dasel/v2"} {
		t.Run(engine, func(t *testing.T) {
			spec := map[string]interface{}{"file": "fixture.json", "key": "version"}
			if engine != "default" {
				spec["engine"] = engine
			}
			resource, err := New(spec)
			if engine == "dasel/v1" || engine == "dasel/v2" {
				require.Error(t, err)
				require.Nil(t, resource)
				return
			}
			require.NoError(t, err)
			require.Equal(t, ENGINEDASEL_V3, resource.engine)
		})
	}
}
