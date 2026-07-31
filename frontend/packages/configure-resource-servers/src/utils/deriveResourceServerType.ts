/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import type {ResourceServerInterface, ResourceServerType} from '../models/resource-server';

/**
 * Derives the Console-only type of a resource server from its interfaces. The backend does not store
 * a type: it stores interfaces, and this is the single place the display rule lives.
 *
 * A resource server with no interfaces defines permissions but has no audience, which is how a Custom
 * resource server is created. One exposing an MCP interface reads as MCP, since that is the notable
 * capability; anything else reads as API. The badge is therefore lossy for a resource server exposing
 * both an API and an MCP interface — the Interfaces tab lists every interface for the exact picture.
 */
export default function deriveResourceServerType(
  interfaces: ResourceServerInterface[] | undefined,
): ResourceServerType {
  if (!interfaces || interfaces.length === 0) {
    return 'CUSTOM';
  }
  if (interfaces.some((rsi) => rsi.type === 'MCP')) {
    return 'MCP';
  }
  return 'API';
}
