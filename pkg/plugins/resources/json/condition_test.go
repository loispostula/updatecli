package json

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
				File:  "testdata/data.json",
				Key:   "firstName",
				Value: "Jack",
			},
			expectedResult: true,
		},
		{
			name: "Default scenario with Dasel v3",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    "firstName",
				Value:  "Jack",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Multiple key scenario",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   "phoneNumbers[0].type",
				Value: "home",
			},
			expectedResult: true,
		},
		{
			name: "Multiple key scenario",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    "phoneNumbers[0].type",
				Value:  "home",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Get last array item successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    "phoneNumbers.filter(type == \"office\").map(number)...",
				Engine: strPtr(ENGINEDASEL_V3),
				Value:  "646 555-4567",
			},
			expectedResult: true,
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   "surname",
				Value: "",
			},
			expectedResult: true,
		},
		{
			name: "Default scenario with Dasel v3",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    "firstName",
				Value:  "Jack",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Nested array item with Dasel v3",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    "phoneNumbers[0].type",
				Value:  "home",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Mismatching value with Dasel v3",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    "firstName",
				Value:  "NotJack",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: false,
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   "doNotExist",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("map key not found"),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   "doNotExist",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("map key not found"),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			got, _, gotErr := j.Condition(context.Background(), "", nil)

			if tt.wantErr {
				require.ErrorContains(t, gotErr, tt.expectedErrorMsg.Error())
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.expectedResult, got)
		})
	}
}
