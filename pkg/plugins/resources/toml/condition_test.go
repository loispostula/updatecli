package toml

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
			name: "Default successful multiple update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  "employees.map(role)...",
			},
			expectedResult: false,
		},
		{
			name: "Default successful multiple update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  "employees.map(role)...",
			},
			expectedResult: false,
		},
		{
			name: "Successful conditional multiple update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  "employees.filter((address ?? \"\") == \"AU\").map(role)...",
			},
			expectedResult: false,
		},
		{
			name: "Successful multiple map update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  "benefits[0].country.filter(country == \"UK\").map(name)...",
			},
			expectedResult: false,
		},
		{
			name: "Default scenario",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   "owner.firstName",
				Value: "Jack",
			},
			expectedResult: true,
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   "owner.surname",
				Value: "",
			},
			expectedResult: true,
		},
		{
			name: "Default scenario with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.firstName",
				Value:  "Jack",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Default scenario with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.firstName",
				Value:  "Jack",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Nested array item with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "servers.beta.role",
				Value:  "backend",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: true,
		},
		{
			name: "Mismatching value with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.firstName",
				Value:  "NotJack",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: false,
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
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
				File:  "testdata/data.toml",
				Key:   "doNotExist...",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("map key not found"),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			toml, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, _, gotErr := toml.Condition(context.Background(), "", nil)

			if tt.wantErr {
				require.ErrorContains(t, gotErr, tt.expectedErrorMsg.Error())
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
