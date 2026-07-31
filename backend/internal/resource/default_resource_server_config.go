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

package resource

import (
	"context"
	"encoding/json"

	"github.com/thunder-id/thunderid/internal/serverconfig"
	"github.com/thunder-id/thunderid/internal/system/resourcedependency"
)

// DefaultResourceServerInterfaceConfig contains the default resource server interface configuration.
// Its owning resource server is the effective default resource server.
type DefaultResourceServerInterfaceConfig struct {
	ResourceServerInterfaceID string `json:"resourceServerInterfaceId" yaml:"resourceServerInterfaceId"`
}

// DefaultResourceServerInterfaceConfigHandler handles default resource server interface configuration.
type DefaultResourceServerInterfaceConfigHandler struct {
	resourceService ResourceServiceInterface
}

// NewDefaultResourceServerInterfaceConfigHandler creates a default resource server interface
// configuration handler.
func NewDefaultResourceServerInterfaceConfigHandler(
	resourceService ResourceServiceInterface,
) *DefaultResourceServerInterfaceConfigHandler {
	if resourceService == nil {
		panic("default resource server interface config handler requires a non-nil resource service")
	}
	return &DefaultResourceServerInterfaceConfigHandler{resourceService: resourceService}
}

// Decode parses a default resource server interface configuration.
func (*DefaultResourceServerInterfaceConfigHandler) Decode(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return DefaultResourceServerInterfaceConfig{}, nil
	}
	var cfg DefaultResourceServerInterfaceConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate validates a default resource server interface configuration.
func (h *DefaultResourceServerInterfaceConfigHandler) Validate(incoming, readOnly, _ any) error {
	cfg, _ := incoming.(DefaultResourceServerInterfaceConfig)
	if ro, ok := readOnly.(DefaultResourceServerInterfaceConfig); ok && ro.ResourceServerInterfaceID != "" {
		return errDeclarativeDefaultLocked
	}
	if cfg.ResourceServerInterfaceID == "" {
		return nil
	}
	_, svcErr := h.resourceService.GetResourceServerInterfaceByID(
		context.Background(), cfg.ResourceServerInterfaceID)
	if svcErr != nil {
		if svcErr.Code == ErrorResourceServerInterfaceNotFound.Code {
			return errUnknownDefaultResourceServerInterface
		}
		return errDefaultResourceServerLookupFailed
	}
	return nil
}

// ReferencedResources implements serverconfig.ReferenceReporter: the section points at one resource
// server interface, so deleting that interface (or the resource server that owns it) is blocked while
// it is the deployment default.
func (*DefaultResourceServerInterfaceConfigHandler) ReferencedResources(
	merged any,
) []serverconfig.ConfigReference {
	cfg, ok := merged.(DefaultResourceServerInterfaceConfig)
	if !ok || cfg.ResourceServerInterfaceID == "" {
		return nil
	}
	return []serverconfig.ConfigReference{{
		ResourceType: resourcedependency.ResourceTypeResourceServerInterface,
		ID:           cfg.ResourceServerInterfaceID,
	}}
}

// Merge combines read-only and writable default resource server interface configurations.
func (*DefaultResourceServerInterfaceConfigHandler) Merge(readOnly, writable any) any {
	if ro, ok := readOnly.(DefaultResourceServerInterfaceConfig); ok && ro.ResourceServerInterfaceID != "" {
		return ro
	}
	if w, ok := writable.(DefaultResourceServerInterfaceConfig); ok {
		return w
	}
	return DefaultResourceServerInterfaceConfig{}
}
