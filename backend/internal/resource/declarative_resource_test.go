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
	"errors"
	"testing"

	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"

	serverconst "github.com/thunder-id/thunderid/internal/system/constants"
	"github.com/thunder-id/thunderid/internal/system/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

// ResourceServerExporterTestSuite tests the resourceServerExporter.
type ResourceServerExporterTestSuite struct {
	suite.Suite
	mockService *ResourceServiceInterfaceMock
	exporter    *resourceServerExporter
	logger      *log.Logger
}

func TestResourceServerExporterTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceServerExporterTestSuite))
}

func (s *ResourceServerExporterTestSuite) SetupTest() {
	s.mockService = NewResourceServiceInterfaceMock(s.T())
	s.exporter = newResourceServerExporterForTest(s.mockService)
	s.logger = log.GetLogger()
}

func (s *ResourceServerExporterTestSuite) TestNewResourceServerExporter() {
	assert.NotNil(s.T(), s.exporter)
}

func (s *ResourceServerExporterTestSuite) TestGetResourceType() {
	assert.Equal(s.T(), "resource_server", s.exporter.GetResourceType())
}

func (s *ResourceServerExporterTestSuite) TestGetParameterizerType() {
	assert.Equal(s.T(), "ResourceServer", s.exporter.GetParameterizerType())
}

func (s *ResourceServerExporterTestSuite) TestGetAllResourceIDs_Success() {
	ctx := context.Background()
	expectedList := &ResourceServerList{
		TotalResults: 2,
		ResourceServers: []providers.ResourceServer{
			{ID: "rs1", Name: "Resource Server 1"},
			{ID: "rs2", Name: "Resource Server 2"},
		},
	}

	s.mockService.EXPECT().GetResourceServerList(ctx, serverconst.MaxPageSize, 0).Return(expectedList, nil)
	s.mockService.EXPECT().IsResourceServerDeclarative("rs1").Return(false)
	s.mockService.EXPECT().IsResourceServerDeclarative("rs2").Return(false)

	ids, err := s.exporter.GetAllResourceIDs(ctx)

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "rs1", ids[0])
	assert.Equal(s.T(), "rs2", ids[1])
}

func (s *ResourceServerExporterTestSuite) TestGetAllResourceIDs_FilterDeclarative() {
	ctx := context.Background()
	expectedList := &ResourceServerList{
		TotalResults: 3,
		ResourceServers: []providers.ResourceServer{
			{ID: "rs1", Name: "Mutable Server"},
			{ID: "rs2", Name: "Declarative Server"},
			{ID: "rs3", Name: "Another Mutable Server"},
		},
	}

	s.mockService.EXPECT().GetResourceServerList(ctx, serverconst.MaxPageSize, 0).Return(expectedList, nil)
	s.mockService.EXPECT().IsResourceServerDeclarative("rs1").Return(false)
	s.mockService.EXPECT().IsResourceServerDeclarative("rs2").Return(true)
	s.mockService.EXPECT().IsResourceServerDeclarative("rs3").Return(false)

	ids, err := s.exporter.GetAllResourceIDs(ctx)

	assert.Nil(s.T(), err)
	assert.Len(s.T(), ids, 2)
	assert.Equal(s.T(), "rs1", ids[0])
	assert.Equal(s.T(), "rs3", ids[1])
}

func (s *ResourceServerExporterTestSuite) TestGetAllResourceIDs_Error() {
	ctx := context.Background()
	expectedError := &tidcommon.ServiceError{
		Code:  "ERR_CODE",
		Error: tidcommon.I18nMessage{DefaultValue: "test error"},
	}

	s.mockService.EXPECT().GetResourceServerList(ctx, serverconst.MaxPageSize, 0).Return(nil, expectedError)

	ids, err := s.exporter.GetAllResourceIDs(ctx)

	assert.Nil(s.T(), ids)
	assert.Equal(s.T(), expectedError, err)
}

func (s *ResourceServerExporterTestSuite) TestGetResourceByID_Success() {
	ctx := context.Background()
	serverID := "rs1"

	server := &providers.ResourceServer{
		ID:          serverID,
		Name:        "Test Server",
		Description: "A test server",
		OUID:        "ou1",
		Delimiter:   ":",
		Interfaces: []providers.ResourceServerInterface{{
			ID:               "rsi1",
			ResourceServerID: serverID,
			Type:             providers.ResourceServerInterfaceTypeAPI,
			Identifier:       "https://api.example.com/test",
		}},
	}

	resources := []providers.Resource{
		{
			ID:           "res1",
			Name:         "Resource 1",
			Handle:       "resource1",
			Description:  "First resource",
			Parent:       nil,
			ParentHandle: "",
			Permission:   "test-server:resource1",
		},
	}

	actions := &ActionList{
		TotalResults: 1,
		Actions: []providers.Action{
			{
				ID:          "act1",
				Name:        "Action 1",
				Handle:      "read",
				Description: "Read action",
				Permission:  "test-server:resource1:read",
				Kind:        providers.ActionKindTool,
			},
		},
	}

	resourceID := "res1"
	s.mockService.EXPECT().GetResourceServer(ctx, serverID).Return(server, nil)
	s.mockService.EXPECT().GetAllResourceList(ctx, serverID).Return(resources, nil)
	s.mockService.EXPECT().
		GetActionList(ctx, serverID, &resourceID, providers.ActionKind(""), serverconst.MaxPageSize, 0).
		Return(actions, nil)

	result, name, err := s.exporter.GetResourceByID(ctx, serverID)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test Server", name)
	assert.NotNil(s.T(), result)

	dto, ok := result.(*providers.ResourceServer)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), serverID, dto.ID)
	assert.Equal(s.T(), "Test Server", dto.Name)
	assert.Len(s.T(), dto.Resources, 1)
	assert.Len(s.T(), dto.Resources[0].Actions, 1)
	assert.Equal(s.T(), providers.ActionKindTool, dto.Resources[0].Actions[0].Kind)
}

func (s *ResourceServerExporterTestSuite) TestGetResourceByID_MCPExportImportRoundTrip() {
	ctx := context.Background()
	serverID := "rs-mcp"

	server := &providers.ResourceServer{
		ID:        serverID,
		Name:      "Booking MCP",
		OUID:      "ou1",
		Delimiter: ":",
		Interfaces: []providers.ResourceServerInterface{{
			ID:               "rsi-mcp",
			ResourceServerID: serverID,
			Type:             providers.ResourceServerInterfaceTypeMCP,
			Identifier:       "https://booking.example.com/mcp",
		}},
	}

	resources := []providers.Resource{
		{
			ID:         "res1",
			Name:       "User Management",
			Handle:     "user-mgmt",
			Parent:     nil,
			Permission: "booking-mcp:user-mgmt",
		},
	}

	actions := &ActionList{
		TotalResults: 1,
		Actions: []providers.Action{
			{
				ID:         "act1",
				Name:       "Create User",
				Handle:     "create_user",
				Permission: "booking-mcp:user-mgmt:create_user",
				Kind:       providers.ActionKindTool,
			},
		},
	}

	resourceID := "res1"
	s.mockService.EXPECT().GetResourceServer(ctx, serverID).Return(server, nil)
	s.mockService.EXPECT().GetAllResourceList(ctx, serverID).Return(resources, nil)
	s.mockService.EXPECT().
		GetActionList(ctx, serverID, &resourceID, providers.ActionKind(""), serverconst.MaxPageSize, 0).
		Return(actions, nil)

	result, _, err := s.exporter.GetResourceByID(ctx, serverID)
	assert.Nil(s.T(), err)

	dto, ok := result.(*providers.ResourceServer)
	assert.True(s.T(), ok)

	// Marshal the exported DTO to YAML, then re-import it. This guards the export->import
	// lossless guarantee: the interfaces must survive so the re-imported resource server keeps
	// its MCP audience identifier.
	yamlBytes, marshalErr := yaml.Marshal(dto)
	assert.NoError(s.T(), marshalErr)

	imported, parseErr := parseToResourceServer(yamlBytes)
	s.Require().NoError(parseErr)
	s.Require().NotNil(imported)
	s.Require().Len(imported.Interfaces, 1)
	assert.Equal(s.T(), providers.ResourceServerInterfaceTypeMCP, imported.Interfaces[0].Type)
	assert.Equal(s.T(), "https://booking.example.com/mcp", imported.Interfaces[0].Identifier)
	assert.Len(s.T(), imported.Resources, 1)
	assert.Len(s.T(), imported.Resources[0].Actions, 1)
	assert.Equal(s.T(), providers.ActionKindTool, imported.Resources[0].Actions[0].Kind)
}

func (s *ResourceServerExporterTestSuite) TestGetResourceByID_ServerNotFound() {
	ctx := context.Background()
	serverID := "rs-nonexistent"

	expectedError := &ErrorResourceServerNotFound

	s.mockService.EXPECT().GetResourceServer(ctx, serverID).Return(nil, expectedError)

	result, name, err := s.exporter.GetResourceByID(ctx, serverID)

	assert.Nil(s.T(), result)
	assert.Equal(s.T(), "", name)
	assert.Equal(s.T(), expectedError, err)
}

func (s *ResourceServerExporterTestSuite) TestValidateResource_Success() {
	dto := &providers.ResourceServer{
		ID:          "rs1",
		Name:        "Test Server",
		Description: "A test server",
		OUID:        "ou1",
	}

	name, err := s.exporter.ValidateResource(context.Background(), dto, "rs1", s.logger)

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Test Server", name)
}

func (s *ResourceServerExporterTestSuite) TestValidateResource_InvalidType() {
	invalidData := "not a resource server dto"

	name, err := s.exporter.ValidateResource(context.Background(), invalidData, "rs1", s.logger)

	assert.Equal(s.T(), "", name)
	assert.NotNil(s.T(), err)
}

func (s *ResourceServerExporterTestSuite) TestValidateResource_EmptyName() {
	dto := &providers.ResourceServer{
		ID:   "rs1",
		Name: "",
		OUID: "ou1",
	}

	name, err := s.exporter.ValidateResource(context.Background(), dto, "rs1", s.logger)

	assert.Equal(s.T(), "", name)
	assert.NotNil(s.T(), err)
}

func (s *ResourceServerExporterTestSuite) TestGetResourceRules() {
	rules := s.exporter.GetResourceRules()

	assert.Nil(s.T(), rules)
}

func TestParseToResourceServer(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
description: "Test description"
ouId: "ou1"
delimiter: ":"
interfaces:
  - id: "rsi1"
    type: "API"
    identifier: "test-server"
resources:
  - name: "Users"
    handle: "users"
    description: "User resources"
    actions:
      - name: "Read"
        handle: "read"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, "rs1", dto.ID)
	assert.Equal(t, "Test Server", dto.Name)
	assert.Len(t, dto.Interfaces, 1)
	assert.Equal(t, "test-server", dto.Interfaces[0].Identifier)
	assert.Equal(t, "ou1", dto.OUID)
	assert.Equal(t, ":", dto.Delimiter)
	assert.Len(t, dto.Resources, 1)
	assert.Equal(t, "users", dto.Resources[0].Handle)
	assert.Len(t, dto.Resources[0].Actions, 1)
}

func TestParseToResourceServer_MissingID(t *testing.T) {
	yamlData := []byte(`
name: "Test Server"
ouId: "ou1"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "ID cannot be empty")
}

func TestParseToResourceServer_MissingName(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
ouId: "ou1"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestParseToResourceServer_InterfaceTypeMCP(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "MCP"
    identifier: "https://api.example.com/rs1"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Len(t, dto.Interfaces, 1)
	assert.Equal(t, providers.ResourceServerInterfaceTypeMCP, dto.Interfaces[0].Type)
}

func TestParseToResourceServer_MissingInterfacesRejected(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "must declare at least one interface")
}

func TestParseToResourceServer_RejectsEmptyInterfaceIdentifier(t *testing.T) {
	// An identifier is an audience, so any non-empty value is accepted; only an empty one is rejected.
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "API"
    identifier: ""
`)

	dto, err := parseToResourceServer(yamlData)

	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "identifier cannot be empty")
}

func TestParseToResourceServer_InvalidInterfaceType(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "BOGUS"
    identifier: "https://api.example.com/rs1"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestParseToResourceServer_ActionKindNotDefaulted(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "MCP"
    identifier: "https://api.example.com/rs1"
resources:
  - name: "Users"
    handle: "users"
    actions:
      - name: "Read"
        handle: "read"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, providers.ActionKind(""), dto.Resources[0].Actions[0].Kind)
}

func TestParseToResourceServer_MCPActionWithKindSucceeds(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "MCP"
    identifier: "https://api.example.com/rs1"
resources:
  - name: "Users"
    handle: "users"
    actions:
      - name: "Read"
        handle: "read"
        kind: "resource"
      - name: "Create"
        handle: "create"
        kind: "tool"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, providers.ActionKindResource, dto.Resources[0].Actions[0].Kind)
	assert.Equal(t, providers.ActionKindTool, dto.Resources[0].Actions[1].Kind)
}

func TestParseToResourceServer_NonMCPActionAllowsKind(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "API"
    identifier: "https://api.example.com/rs1"
resources:
  - name: "Users"
    handle: "users"
    actions:
      - name: "Read"
        handle: "read"
        kind: "tool"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, providers.ActionKindTool, dto.Resources[0].Actions[0].Kind)
}

func TestParseToResourceServer_NonMCPActionNoKindStaysEmpty(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "API"
    identifier: "https://api.example.com/rs1"
resources:
  - name: "Users"
    handle: "users"
    actions:
      - name: "Read"
        handle: "read"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.NoError(t, err)
	assert.NotNil(t, dto)
	assert.Equal(t, providers.ActionKind(""), dto.Resources[0].Actions[0].Kind)
}

func TestParseToResourceServer_ActionInvalidKindRejected(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "MCP"
    identifier: "https://api.example.com/rs1"
resources:
  - name: "Users"
    handle: "users"
    actions:
      - name: "Read"
        handle: "read"
        kind: "prompt"
`)

	dto, err := parseToResourceServer(yamlData)

	assert.Error(t, err)
	assert.Nil(t, dto)
	assert.Contains(t, err.Error(), "read")
	assert.Contains(t, err.Error(), "invalid kind")
}

func TestParseAndValidateResourceServerWrapper_TypeMCP(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
handle: "test-api"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "MCP"
    identifier: "https://api.example.com/rs1"
`)

	parser := parseAndValidateResourceServerWrapper(nil)
	result, err := parser(yamlData)

	assert.NoError(t, err)
	rs, ok := result.(*providers.ResourceServer)
	assert.True(t, ok)
	assert.Len(t, rs.Interfaces, 1)
	assert.Equal(t, providers.ResourceServerInterfaceTypeMCP, rs.Interfaces[0].Type)
}

func TestParseAndValidateResourceServerWrapper_InvalidType(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
handle: "test-api"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "BOGUS"
    identifier: "https://api.example.com/rs1"
`)

	parser := parseAndValidateResourceServerWrapper(nil)
	result, err := parser(yamlData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestBuildPermissionString(t *testing.T) {
	resourceHandleMap := map[string]*providers.Resource{
		"users": {
			Handle:       "users",
			Parent:       nil,
			ParentHandle: "",
		},
		"profile": {
			Handle:       "profile",
			Parent:       nil,
			ParentHandle: "users",
		},
	}

	tests := []struct {
		name      string
		resource  *providers.Resource
		delimiter string
		expected  string
	}{
		{
			name: "root resource",
			resource: &providers.Resource{
				Handle:       "users",
				ParentHandle: "",
			},
			delimiter: ":",
			expected:  "users",
		},
		{
			name: "nested resource",
			resource: &providers.Resource{
				Handle:       "profile",
				ParentHandle: "users",
			},
			delimiter: ":",
			expected:  "users:profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildPermissionString(tt.resource, resourceHandleMap, tt.delimiter)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessResourceServer_SetsPermissionsAndDelimiter(t *testing.T) {
	rs := &providers.ResourceServer{
		ID:   "rs1",
		Name: "Test Server",
		OUID: "ou1",
		Resources: []providers.Resource{
			{
				Name:   "Users",
				Handle: "users",
				Actions: []providers.Action{
					{Name: "Read", Handle: "read"},
				},
			},
			{
				Name:         "Profile",
				Handle:       "profile",
				ParentHandle: "users",
			},
		},
	}

	err := ProcessResourceServer(rs)

	assert.NoError(t, err)
	assert.Equal(t, ":", rs.Delimiter)
	assert.Equal(t, "users", rs.Resources[0].Permission)
	assert.Equal(t, "users:read", rs.Resources[0].Actions[0].Permission)
	assert.Equal(t, "users:profile", rs.Resources[1].Permission)
}

func TestProcessResourceServer_DuplicateHandle(t *testing.T) {
	rs := &providers.ResourceServer{
		ID:   "rs1",
		Name: "Test Server",
		OUID: "ou1",
		Resources: []providers.Resource{
			{Name: "Users", Handle: "users"},
			{Name: "Users Duplicate", Handle: "users"},
		},
	}

	err := ProcessResourceServer(rs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate resource handle")
}

func TestProcessResourceServer_ActionCollidesWithGroupPermission(t *testing.T) {
	// A group (RESOURCE) nested under "ops" shares its handle with an action nested under the same
	// "ops" group: both derive "ops:deploy".
	rs := &providers.ResourceServer{
		ID:   "rs-mcp",
		Name: "Booking MCP",
		OUID: "ou1",
		Resources: []providers.Resource{
			{
				Name:   "Ops",
				Handle: "ops",
				Actions: []providers.Action{
					{Name: "Deploy", Handle: "deploy", Kind: providers.ActionKindTool},
				},
			},
			{
				Name:         "Deploy Group",
				Handle:       "deploy",
				ParentHandle: "ops",
			},
		},
	}

	err := ProcessResourceServer(rs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate permission")
	assert.Contains(t, err.Error(), "ops:deploy")
}

func TestProcessResourceServer_ActionCollidesWithNestedGroupPermission(t *testing.T) {
	// An action nested under group "a" derives "a:b". A child group "b" nested under "a" derives the
	// same "a:b". The cross-entity collision is caught even though the two collide via different
	// nesting paths rather than at the same level.
	rs := &providers.ResourceServer{
		ID:   "rs-mcp",
		Name: "Booking MCP",
		OUID: "ou1",
		Resources: []providers.Resource{
			{
				Name:   "Group A",
				Handle: "a",
				Actions: []providers.Action{
					{Name: "Tool B", Handle: "b", Kind: providers.ActionKindTool},
				},
			},
			{
				Name:         "Group B",
				Handle:       "b",
				ParentHandle: "a",
			},
		},
	}

	err := ProcessResourceServer(rs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate permission")
	assert.Contains(t, err.Error(), "a:b")
}

func TestProcessResourceServer_NoCollisionSucceeds(t *testing.T) {
	rs := &providers.ResourceServer{
		ID:   "rs-mcp",
		Name: "Booking MCP",
		OUID: "ou1",
		Resources: []providers.Resource{
			{
				Name:   "Ops",
				Handle: "ops",
				Actions: []providers.Action{
					{Name: "Deploy", Handle: "deploy", Kind: providers.ActionKindTool},
					{Name: "Status", Handle: "status", Kind: providers.ActionKindResource},
				},
			},
			{
				Name:   "Users",
				Handle: "users",
				Actions: []providers.Action{
					{Name: "Create", Handle: "create", Kind: providers.ActionKindTool},
				},
			},
		},
	}

	err := ProcessResourceServer(rs)

	assert.NoError(t, err)
	assert.Equal(t, "ops:deploy", rs.Resources[0].Actions[0].Permission)
	assert.Equal(t, "ops:status", rs.Resources[0].Actions[1].Permission)
	assert.Equal(t, "users:create", rs.Resources[1].Actions[0].Permission)
}

func TestProcessResourceServer_CollisionCheckIsInterfaceIndependent(t *testing.T) {
	// Permissions belong to the resource server, not to any interface, so a colliding structure is
	// rejected regardless of the interfaces the resource server exposes.
	rs := &providers.ResourceServer{
		ID:   "rs-api",
		Name: "Booking API",
		OUID: "ou1",
		Interfaces: []providers.ResourceServerInterface{{
			ID:         "rsi1",
			Type:       providers.ResourceServerInterfaceTypeAPI,
			Identifier: "https://booking.example.com/api",
		}},
		Resources: []providers.Resource{
			{
				Name:   "Ops",
				Handle: "ops",
				Actions: []providers.Action{
					{Name: "Deploy", Handle: "deploy"},
				},
			},
			{
				Name:         "Deploy Group",
				Handle:       "deploy",
				ParentHandle: "ops",
			},
		},
	}

	err := ProcessResourceServer(rs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate permission")
	assert.Contains(t, err.Error(), "ops:deploy")
}

func TestProcessResource_SetsPermissions(t *testing.T) {
	root := &providers.Resource{Handle: "root"}
	resource := &providers.Resource{
		Handle:       "child",
		ParentHandle: "root",
		Actions: []providers.Action{
			{Name: "Read", Handle: "read"},
		},
	}
	resourceHandleMap := map[string]*providers.Resource{
		"root":  root,
		"child": resource,
	}

	err := processResource(resource, resourceHandleMap, ":")

	assert.NoError(t, err)
	assert.Equal(t, "root:child", resource.Permission)
	assert.Equal(t, "root:child:read", resource.Actions[0].Permission)
}

func TestProcessResource_MissingParent(t *testing.T) {
	resource := &providers.Resource{Handle: "child", ParentHandle: "missing"}
	resourceHandleMap := map[string]*providers.Resource{}

	err := processResource(resource, resourceHandleMap, ":")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent resource handle")
}

func TestParseAndValidateResourceServerWrapper_Success(t *testing.T) {
	yamlData := []byte(`
id: "rs1"
name: "Test Server"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "API"
    identifier: "api"
resources:
  - name: "Users"
    handle: "users"
    actions:
      - name: "Read"
        handle: "read"
  - name: "Profile"
    handle: "profile"
    parent: "users"
`)

	parser := parseAndValidateResourceServerWrapper(nil)
	result, err := parser(yamlData)

	assert.NoError(t, err)
	rs, ok := result.(*providers.ResourceServer)
	assert.True(t, ok)
	assert.Equal(t, "users", rs.Resources[0].Permission)
	assert.Equal(t, "users:read", rs.Resources[0].Actions[0].Permission)
	assert.Equal(t, "users:profile", rs.Resources[1].Permission)
}

func TestParseAndValidateResourceServerWrapper_MCPPermissionCollisionRejected(t *testing.T) {
	yamlData := []byte(`
id: "rs-mcp"
name: "Booking MCP"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "MCP"
    identifier: "https://api.example.com/rs1"
resources:
  - name: "Ops"
    handle: "ops"
    actions:
      - name: "Deploy"
        handle: "deploy"
        kind: "tool"
  - name: "Deploy Group"
    handle: "deploy"
    parent: "ops"
`)

	parser := parseAndValidateResourceServerWrapper(nil)
	result, err := parser(yamlData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "duplicate permission")
	assert.Contains(t, err.Error(), "ops:deploy")
}

func TestParseAndValidateResourceServerWrapper_MCPCleanSucceeds(t *testing.T) {
	yamlData := []byte(`
id: "rs-mcp"
name: "Booking MCP"
ouId: "ou1"
interfaces:
  - id: "rsi1"
    type: "MCP"
    identifier: "https://api.example.com/rs1"
resources:
  - name: "Ops"
    handle: "ops"
    actions:
      - name: "Deploy"
        handle: "deploy"
        kind: "tool"
      - name: "Status"
        handle: "status"
        kind: "resource"
`)

	parser := parseAndValidateResourceServerWrapper(nil)
	result, err := parser(yamlData)

	assert.NoError(t, err)
	rs, ok := result.(*providers.ResourceServer)
	assert.True(t, ok)
	assert.Equal(t, "ops:deploy", rs.Resources[0].Actions[0].Permission)
	assert.Equal(t, "ops:status", rs.Resources[0].Actions[1].Permission)
}

func TestParseAndValidateResourceServerWrapper_InvalidYAML(t *testing.T) {
	yamlData := []byte(`::invalid`)

	parser := parseAndValidateResourceServerWrapper(nil)
	result, err := parser(yamlData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestValidateResourceServerWrapper_InvalidType(t *testing.T) {
	err := validateResourceServerWrapper("invalid", newResourceStoreInterfaceMock(t), nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestValidateResourceServerWrapper_EmptyName(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)

	err := validateResourceServerWrapper(&providers.ResourceServer{ID: "rs1"}, fileStore, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

// testInterfaces returns a single valid API interface for declarative validation tests.
func testInterfaces() []providers.ResourceServerInterface {
	return []providers.ResourceServerInterface{{
		ID:         "rsi1",
		Type:       providers.ResourceServerInterfaceTypeAPI,
		Identifier: "https://api.example.com/test",
	}}
}

func TestValidateResourceServerWrapper_MissingInterfaces(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)

	err := validateResourceServerWrapper(&providers.ResourceServer{ID: "rs1", Name: "Server"}, fileStore, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must declare at least one interface")
}

func TestValidateResourceServerWrapper_DuplicateInFileStore(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs1").Return(providers.ResourceServer{ID: "rs1"}, nil)

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs1", Name: "Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate resource server ID")
}

func TestValidateResourceServerWrapper_FileStoreError(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs1").Return(providers.ResourceServer{}, errors.New("file error"))

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs1", Name: "Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check")
}

func TestValidateResourceServerWrapper_DuplicateInDBStore(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	dbStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs1").
		Return(providers.ResourceServer{}, errResourceServerNotFound)
	dbStore.On("GetResourceServer", mock.Anything, "rs1").Return(providers.ResourceServer{ID: "rs1"}, nil)

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs1", Name: "Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, dbStore, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database store")
}

func TestValidateResourceServerWrapper_DBStoreError(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	dbStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs1").
		Return(providers.ResourceServer{}, errResourceServerNotFound)
	dbStore.On("GetResourceServer", mock.Anything, "rs1").Return(providers.ResourceServer{}, errors.New("db error"))

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs1", Name: "Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, dbStore, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database store")
}

func TestValidateResourceServerWrapper_Success(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs1").
		Return(providers.ResourceServer{}, errResourceServerNotFound)
	// No other resource server claims this interface ID or identifier.
	fileStore.On("GetResourceServerInterface", mock.Anything, "rsi1").
		Return(providers.ResourceServerInterface{}, errResourceServerInterfaceNotFound)
	fileStore.On("GetResourceServerInterfaceByIdentifier", mock.Anything, "https://api.example.com/test").
		Return(providers.ResourceServerInterface{}, errResourceServerInterfaceNotFound)

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs1", Name: "Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, nil, nil)

	assert.NoError(t, err)
}

// A second declarative resource server must not claim an identifier another one already exposes:
// identifier resolution returns the first match, so the audience would depend on store ordering.
func TestValidateResourceServerWrapper_DuplicateInterfaceIdentifierAcrossServers(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs2").
		Return(providers.ResourceServer{}, errResourceServerNotFound)
	fileStore.On("GetResourceServerInterface", mock.Anything, "rsi1").
		Return(providers.ResourceServerInterface{}, errResourceServerInterfaceNotFound)
	fileStore.On("GetResourceServerInterfaceByIdentifier", mock.Anything, "https://api.example.com/test").
		Return(providers.ResourceServerInterface{
			ID:               "rsi-other",
			ResourceServerID: "rs1",
			Identifier:       "https://api.example.com/test",
		}, nil)

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs2", Name: "Second Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate interface identifier")
	assert.Contains(t, err.Error(), "rs1")
}

// An interface ID already used by another resource server is rejected too, since the default
// interface configuration references interfaces by ID.
func TestValidateResourceServerWrapper_DuplicateInterfaceIDAcrossServers(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs2").
		Return(providers.ResourceServer{}, errResourceServerNotFound)
	fileStore.On("GetResourceServerInterface", mock.Anything, "rsi1").
		Return(providers.ResourceServerInterface{ID: "rsi1", ResourceServerID: "rs1"}, nil)

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs2", Name: "Second Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate interface ID")
}

// In composite mode a declarative interface must not collide with a database-backed one either.
func TestValidateResourceServerWrapper_DuplicateInterfaceIdentifierInDBStore(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	dbStore := newResourceStoreInterfaceMock(t)
	fileStore.On("GetResourceServer", mock.Anything, "rs2").
		Return(providers.ResourceServer{}, errResourceServerNotFound)
	dbStore.On("GetResourceServer", mock.Anything, "rs2").
		Return(providers.ResourceServer{}, errResourceServerNotFound)
	fileStore.On("GetResourceServerInterface", mock.Anything, "rsi1").
		Return(providers.ResourceServerInterface{}, errResourceServerInterfaceNotFound).Maybe()
	fileStore.On("GetResourceServerInterfaceByIdentifier", mock.Anything, "https://api.example.com/test").
		Return(providers.ResourceServerInterface{}, errResourceServerInterfaceNotFound).Maybe()
	dbStore.On("GetResourceServerInterface", mock.Anything, "rsi1").
		Return(providers.ResourceServerInterface{}, errResourceServerInterfaceNotFound).Maybe()
	dbStore.On("GetResourceServerInterfaceByIdentifier", mock.Anything, "https://api.example.com/test").
		Return(providers.ResourceServerInterface{
			ID:               "rsi-db",
			ResourceServerID: "rs-db",
			Identifier:       "https://api.example.com/test",
		}, nil).Maybe()

	err := validateResourceServerWrapper(
		&providers.ResourceServer{ID: "rs2", Name: "Second Server", OUID: "ou1", Interfaces: testInterfaces()},
		fileStore, dbStore, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate interface identifier")
}

func TestLoadDeclarativeResources_CompositeFileStoreTypeError(t *testing.T) {
	fileStore := newResourceStoreInterfaceMock(t)
	dbStore := newResourceStoreInterfaceMock(t)
	compositeStore := newCompositeResourceStore(fileStore, dbStore)

	err := loadDeclarativeResources(compositeStore, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assert fileStore")
}

func TestLoadDeclarativeResources_InvalidStoreType(t *testing.T) {
	invalidStore := newResourceStoreInterfaceMock(t)

	err := loadDeclarativeResources(invalidStore, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid store type")
}
