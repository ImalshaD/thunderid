/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package serverconfig

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/thunder-id/thunderid/internal/system/resourcedependency"
)

// referencingHandler is a config handler whose section points at one resource.
type referencingHandler struct {
	referencedID string
}

func (*referencingHandler) Decode(json.RawMessage) (any, error) { return nil, nil }
func (*referencingHandler) Validate(_, _, _ any) error          { return nil }
func (*referencingHandler) Merge(_, _ any) any                  { return nil }

func (h *referencingHandler) ReferencedResources(_ any) []ConfigReference {
	return []ConfigReference{{
		ResourceType: resourcedependency.ResourceTypeResourceServerInterface,
		ID:           h.referencedID,
	}}
}

// plainHandler holds no references and must be skipped.
type plainHandler struct{}

func (plainHandler) Decode(json.RawMessage) (any, error) { return nil, nil }
func (plainHandler) Validate(_, _, _ any) error          { return nil }
func (plainHandler) Merge(_, _ any) any                  { return nil }

// newTestService wires the handlers over an empty store, so every section reads as unset.
func newTestService(t *testing.T, handlers map[ConfigName]ServerConfigHandlerInterface) ServerConfigService {
	store := newServerConfigStoreInterfaceMock(t)
	store.EXPECT().GetServerConfig(mock.Anything, mock.Anything).Return(storeLayers{}, nil).Maybe()
	return newServerConfigService(store, handlers)
}

func TestGetResourceDependencies_ReportsReferencedResource(t *testing.T) {
	svc := newTestService(t, map[ConfigName]ServerConfigHandlerInterface{
		ConfigNameDefaultResourceServerInterface: &referencingHandler{referencedID: "rsi-1"},
	})

	deps, err := svc.GetResourceDependencies(
		context.Background(), resourcedependency.ResourceTypeResourceServerInterface, "rsi-1")

	require.NoError(t, err)
	require.Len(t, deps, 1)
	assert.Equal(t, resourcedependency.ResourceTypeServerConfig, deps[0].ResourceType)
	assert.Equal(t, string(ConfigNameDefaultResourceServerInterface), deps[0].ID)
	// A restrict behavior is what makes the delete fail rather than cascade.
	assert.Equal(t, resourcedependency.BehaviorRestrict, deps[0].BehaviorOnDelete)
}

func TestGetResourceDependencies_IgnoresOtherResources(t *testing.T) {
	svc := newTestService(t, map[ConfigName]ServerConfigHandlerInterface{
		ConfigNameDefaultResourceServerInterface: &referencingHandler{referencedID: "rsi-1"},
	})

	deps, err := svc.GetResourceDependencies(
		context.Background(), resourcedependency.ResourceTypeResourceServerInterface, "rsi-other")

	require.NoError(t, err)
	assert.Empty(t, deps)
}

func TestGetResourceDependencies_SkipsHandlersWithoutReferences(t *testing.T) {
	svc := newTestService(t, map[ConfigName]ServerConfigHandlerInterface{
		ConfigNameCORS: plainHandler{},
	})

	deps, err := svc.GetResourceDependencies(
		context.Background(), resourcedependency.ResourceTypeResourceServerInterface, "rsi-1")

	require.NoError(t, err)
	assert.Empty(t, deps)
}
