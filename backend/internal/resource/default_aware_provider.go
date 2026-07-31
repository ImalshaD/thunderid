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

	"github.com/thunder-id/thunderid/internal/serverconfig"
	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

// defaultAwareResourceServerProvider decorates a providers.ResourceServerProvider so that an empty
// identifier resolves the deployment's configured default resource server. The default resource server
// is a server-side policy: the authentication engine and OAuth layers depend only on
// providers.ResourceServerProvider and never see the server-config store.
type defaultAwareResourceServerProvider struct {
	ResourceServiceInterface
	serverConfigService serverconfig.ServerConfigService
}

var _ providers.ResourceServerProvider = (*defaultAwareResourceServerProvider)(nil)

// NewDefaultAwareResourceServerProvider wraps resourceService so that GetResourceServerByIdentifier
// resolves the resource server owning the configured default interface when the identifier is empty.
// resourceService and serverConfigService must both be non-nil.
func NewDefaultAwareResourceServerProvider(
	resourceService ResourceServiceInterface,
	serverConfigService serverconfig.ServerConfigService,
) providers.ResourceServerProvider {
	if resourceService == nil {
		panic("default-aware resource server provider requires a non-nil resource service")
	}
	if serverConfigService == nil {
		panic("default-aware resource server provider requires a non-nil server config service")
	}
	return &defaultAwareResourceServerProvider{
		ResourceServiceInterface: resourceService,
		serverConfigService:      serverConfigService,
	}
}

// GetResourceServerByIdentifier resolves an explicit identifier through the wrapped provider. When the
// identifier is empty it resolves the deployment's configured default interface: a client error
// when no default is configured (or the merged config is malformed), and a server error when the
// configuration cannot be read.
//
// Either way the returned resource server carries exactly the interface that selected it, so callers
// bind the access-token audience to that interface identifier while authorizing against the resource
// server itself.
func (p *defaultAwareResourceServerProvider) GetResourceServerByIdentifier(
	ctx context.Context, identifier string,
) (*providers.ResourceServer, *tidcommon.ServiceError) {
	if identifier != "" {
		return p.ResourceServiceInterface.GetResourceServerByIdentifier(ctx, identifier)
	}
	merged, svcErr := p.serverConfigService.GetMergedConfig(
		ctx, string(serverconfig.ConfigNameDefaultResourceServerInterface))
	if svcErr != nil {
		return nil, svcErr
	}
	cfg, _ := merged.(DefaultResourceServerInterfaceConfig)
	if cfg.ResourceServerInterfaceID == "" {
		return nil, &ErrorResourceServerNotFound
	}
	defaultInterface, svcErr := p.ResourceServiceInterface.GetResourceServerInterfaceByID(
		ctx, cfg.ResourceServerInterfaceID)
	if svcErr != nil {
		return nil, svcErr
	}
	resourceServer, svcErr := p.ResourceServiceInterface.GetResourceServer(
		ctx, defaultInterface.ResourceServerID)
	if svcErr != nil {
		return nil, svcErr
	}

	narrowed := *resourceServer
	narrowed.Interfaces = []providers.ResourceServerInterface{*defaultInterface}
	return &narrowed, nil
}
