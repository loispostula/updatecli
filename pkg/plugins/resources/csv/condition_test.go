package csv

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Default scenario",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   "$this[0].firstname",
				Value: "John",
			},
			expectedResult: true,
		},
		{
			name: "Deprecated Multiple query scenario",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   "map(firstname)...",
				Value: "John",
			},
			expectedResult: false,
		},
		{
			name: "Query scenario",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   "map(firstname)...",
				Value: "John",
			},
			expectedResult: false,
		},
		{
			name: "Multiple query scenario",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   "map(surname)...",
				Value: "",
			},
			expectedResult: false,
		},
		{
			name: "Default scenario 2",
			spec: Spec{
				File:  "testdata/data.2.csv",
				Key:   "$this[0].firstname",
				Comma: ';',
				Value: "John",
			},
			expectedResult: true,
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   "$this[0].surname",
				Value: "",
			},
			expectedResult: true,
		},
		{
			name: "Default scenario with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].firstname",
				Value:  "John",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Default scenario with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].firstname",
				Value:  "John",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Mismatching value with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].firstname",
				Value:  "NotJohn",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: false,
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   "$this[0].doNotExist",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("map key not found"),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.spec)

			require.NoError(t, err)

			got, _, gotErr := c.Condition(context.Background(), "", nil)

			if tt.wantErr {
				require.Error(t, gotErr)
				require.ErrorContains(t, gotErr, tt.expectedErrorMsg.Error())
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedResult, got)
		})
	}
}
