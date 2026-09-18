package yaml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectLegacyYamlPath(t *testing.T) {
	for _, key := range []string{`image`, `image.tag`, `image\.tag`, `$.image\.tag`} {
		t.Run(key, func(t *testing.T) {
			_, err := New(Spec{File: "test.yaml", Key: key})
			require.ErrorContains(t, err, "legacy YAML key")
			_, err = New(Spec{File: "test.yaml", Keys: []string{key}})
			require.ErrorContains(t, err, "legacy YAML key")
		})
	}
}
