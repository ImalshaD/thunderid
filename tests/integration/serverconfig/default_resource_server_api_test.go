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
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

type DefaultResourceServerInterfaceAPITestSuite struct {
	suite.Suite
	adminClient *http.Client
}

func TestDefaultResourceServerInterfaceAPITestSuite(t *testing.T) {
	suite.Run(t, new(DefaultResourceServerInterfaceAPITestSuite))
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) SetupSuite() {
	suite.adminClient = testutils.GetHTTPClient()
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) SetupTest()     { suite.clear() }
func (suite *DefaultResourceServerInterfaceAPITestSuite) TearDownSuite() { suite.clear() }

func (suite *DefaultResourceServerInterfaceAPITestSuite) TestListIncludesSection() {
	status, body := suite.get(serverConfigURL)
	suite.Require().Equal(http.StatusOK, status)

	var names []string
	suite.Require().NoError(json.Unmarshal(body, &names))
	suite.Contains(names, "defaultResourceServerInterface")
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) TestPutExistingInterfaceIDPersistsAndReads() {
	status, _ := suite.put(`{"resourceServerInterfaceId":"` + systemResourceServerInterfaceID + `"}`)
	suite.Require().Equal(http.StatusOK, status)

	layers := suite.getLayers()
	suite.Equal(systemResourceServerInterfaceID, layers.Writable.ResourceServerInterfaceID)
	suite.Equal(systemResourceServerInterfaceID, layers.Merged.ResourceServerInterfaceID)
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) TestPutUnknownInterfaceIDReturns400AndDoesNotPersist() {
	status, body := suite.put(`{"resourceServerInterfaceId":"00000000-0000-0000-0000-000000000000"}`)
	suite.Equal(http.StatusBadRequest, status)
	suite.Equal("SCF-1003", suite.errorCode(body))

	suite.Empty(suite.getLayers().Writable.ResourceServerInterfaceID)
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) TestClearIsAccepted() {
	suite.Require().Equal(http.StatusOK, suite.mustPut(`{"resourceServerInterfaceId":"`+systemResourceServerInterfaceID+`"}`))

	status, _ := suite.put(`{"resourceServerInterfaceId":""}`)
	suite.Require().Equal(http.StatusOK, status)
	suite.Empty(suite.getLayers().Writable.ResourceServerInterfaceID)
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) clear() {
	suite.Require().Equal(http.StatusOK, suite.mustPut(`{"resourceServerInterfaceId":""}`))
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) mustPut(body string) int {
	status, _ := suite.put(body)
	return status
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) put(body string) (int, []byte) {
	req, err := http.NewRequest(http.MethodPut, defaultResourceServerInterfaceConfigURL, strings.NewReader(body))
	suite.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.adminClient.Do(req)
	suite.Require().NoError(err)
	defer closeBodyQuietly(suite.T(), resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	return resp.StatusCode, respBody
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) get(url string) (int, []byte) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	suite.Require().NoError(err)

	resp, err := suite.adminClient.Do(req)
	suite.Require().NoError(err)
	defer closeBodyQuietly(suite.T(), resp.Body)

	body, err := io.ReadAll(resp.Body)
	suite.Require().NoError(err)
	return resp.StatusCode, body
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) getLayers() defaultResourceServerInterfaceLayers {
	status, body := suite.get(defaultResourceServerInterfaceConfigURL)
	suite.Require().Equal(http.StatusOK, status)
	var layers defaultResourceServerInterfaceLayers
	suite.Require().NoError(json.Unmarshal(body, &layers))
	return layers
}

func (suite *DefaultResourceServerInterfaceAPITestSuite) errorCode(body []byte) string {
	var errResp apiErrorResponse
	suite.Require().NoError(json.Unmarshal(body, &errResp))
	return errResp.Code
}
