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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
)

type DefaultResourceServerInterfaceConfigHandlerTestSuite struct {
	suite.Suite
}

func TestDefaultResourceServerInterfaceConfigHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultResourceServerInterfaceConfigHandlerTestSuite))
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) marshal(v any) string {
	out, err := json.Marshal(v)
	suite.Require().NoError(err)
	return string(out)
}

type defaultInterfaceConfig = DefaultResourceServerInterfaceConfig

type defaultInterfaceConfigHandler = DefaultResourceServerInterfaceConfigHandler

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) handler() *defaultInterfaceConfigHandler {
	return NewDefaultResourceServerInterfaceConfigHandler(NewResourceServiceInterfaceMock(suite.T()))
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestDecodeEmptyYieldsEmptyConfig() {
	decoded, err := suite.handler().Decode(json.RawMessage(nil))
	suite.Require().NoError(err)
	assert.JSONEq(suite.T(), `{"resourceServerInterfaceId":""}`, suite.marshal(decoded))
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestDecodeValidObject() {
	decoded, err := suite.handler().Decode(json.RawMessage(`{"resourceServerInterfaceId":"abc"}`))
	suite.Require().NoError(err)
	assert.Equal(suite.T(), DefaultResourceServerInterfaceConfig{ResourceServerInterfaceID: "abc"}, decoded)
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestDecodeMalformedJSON() {
	_, err := suite.handler().Decode(json.RawMessage(`{not json`))
	assert.Error(suite.T(), err)
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestValidateUnsetAccepted() {
	h := NewDefaultResourceServerInterfaceConfigHandler(NewResourceServiceInterfaceMock(suite.T()))
	assert.NoError(suite.T(), h.Validate(defaultInterfaceConfig{}, nil, nil))
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestValidateKnownIDAccepted() {
	mockSvc := NewResourceServiceInterfaceMock(suite.T())
	mockSvc.EXPECT().GetResourceServerInterfaceByID(mock.Anything, "rsi-1").
		Return(&providers.ResourceServerInterface{ID: "rsi-1", ResourceServerID: "rs-1"}, nil)
	h := NewDefaultResourceServerInterfaceConfigHandler(mockSvc)
	assert.NoError(suite.T(), h.Validate(defaultInterfaceConfig{ResourceServerInterfaceID: "rsi-1"}, nil, nil))
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestValidateUnknownIDRejected() {
	mockSvc := NewResourceServiceInterfaceMock(suite.T())
	mockSvc.EXPECT().GetResourceServerInterfaceByID(mock.Anything, "missing").
		Return(nil, &ErrorResourceServerInterfaceNotFound)
	h := NewDefaultResourceServerInterfaceConfigHandler(mockSvc)
	assert.Error(suite.T(), h.Validate(defaultInterfaceConfig{ResourceServerInterfaceID: "missing"}, nil, nil))
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestValidateInternalErrorRejected() {
	mockSvc := NewResourceServiceInterfaceMock(suite.T())
	mockSvc.EXPECT().GetResourceServerInterfaceByID(mock.Anything, "rsi-1").
		Return(nil, &tidcommon.InternalServerError)
	h := NewDefaultResourceServerInterfaceConfigHandler(mockSvc)
	err := h.Validate(defaultInterfaceConfig{ResourceServerInterfaceID: "rsi-1"}, nil, nil)
	assert.ErrorIs(suite.T(), err, errDefaultResourceServerLookupFailed)
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestValidateRejectsWriteWhenDeclarativeSet() {
	h := NewDefaultResourceServerInterfaceConfigHandler(NewResourceServiceInterfaceMock(suite.T()))
	err := h.Validate(
		defaultInterfaceConfig{ResourceServerInterfaceID: "rsi-2"},
		defaultInterfaceConfig{ResourceServerInterfaceID: "rsi-1"},
		defaultInterfaceConfig{},
	)
	assert.Error(suite.T(), err)
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestValidateRejectsClearWhenDeclarativeSet() {
	h := NewDefaultResourceServerInterfaceConfigHandler(NewResourceServiceInterfaceMock(suite.T()))
	err := h.Validate(
		defaultInterfaceConfig{},
		defaultInterfaceConfig{ResourceServerInterfaceID: "rsi-1"},
		defaultInterfaceConfig{},
	)
	assert.Error(suite.T(), err)
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestMergeReadOnlyWins() {
	merged := suite.handler().Merge(
		defaultInterfaceConfig{ResourceServerInterfaceID: "ro"},
		defaultInterfaceConfig{ResourceServerInterfaceID: "w"},
	)
	assert.Equal(suite.T(), defaultInterfaceConfig{ResourceServerInterfaceID: "ro"}, merged)
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestMergeFallsBackToWritable() {
	merged := suite.handler().
		Merge(defaultInterfaceConfig{}, defaultInterfaceConfig{ResourceServerInterfaceID: "w"})
	assert.Equal(suite.T(), defaultInterfaceConfig{ResourceServerInterfaceID: "w"}, merged)
}

func (suite *DefaultResourceServerInterfaceConfigHandlerTestSuite) TestMergeBothEmpty() {
	merged := suite.handler().
		Merge(defaultInterfaceConfig{}, defaultInterfaceConfig{})
	assert.Equal(suite.T(), defaultInterfaceConfig{}, merged)
}
