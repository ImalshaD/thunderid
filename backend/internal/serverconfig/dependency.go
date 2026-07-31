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
	"fmt"

	"github.com/thunder-id/thunderid/internal/system/resourcedependency"
)

// ConfigReference is a resource a configuration section points at.
type ConfigReference struct {
	ResourceType string
	ID           string
}

// ReferenceReporter is an optional interface a configuration handler may implement to report the
// resources its section references. Handlers own their value type, so reporting is inverted rather
// than having this package decode every section: a section referencing a resource elsewhere in the
// product would otherwise force an import of that package.
type ReferenceReporter interface {
	ReferencedResources(merged any) []ConfigReference
}

// GetResourceDependencies implements resourcedependency.Provider. A configuration section that
// references a resource blocks its deletion, since removing the target would leave the section
// pointing at something that no longer exists.
func (s *serverConfigService) GetResourceDependencies(
	ctx context.Context, resourceType, id string,
) ([]resourcedependency.ResourceDependency, error) {
	dependencies := make([]resourcedependency.ResourceDependency, 0)

	for name, handler := range s.handlers {
		reporter, ok := handler.(ReferenceReporter)
		if !ok {
			continue
		}

		merged, svcErr := s.GetMergedConfig(ctx, string(name))
		if svcErr != nil {
			return nil, fmt.Errorf("failed to read server configuration %q: %s", name, svcErr.Error.DefaultValue)
		}

		for _, ref := range reporter.ReferencedResources(merged) {
			if ref.ResourceType != resourceType || ref.ID != id {
				continue
			}
			dependencies = append(dependencies, resourcedependency.ResourceDependency{
				ResourceType:     resourcedependency.ResourceTypeServerConfig,
				ID:               string(name),
				DisplayName:      string(name),
				BehaviorOnDelete: resourcedependency.BehaviorRestrict,
			})
		}
	}

	return dependencies, nil
}
