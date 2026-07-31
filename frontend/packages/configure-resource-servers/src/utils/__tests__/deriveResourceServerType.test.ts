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

import {describe, expect, it} from 'vitest';
import type {ResourceServerInterface} from '../../models/resource-server';
import deriveResourceServerType from '../deriveResourceServerType';

const iface = (type: 'API' | 'MCP', id: string): ResourceServerInterface => ({
  id,
  type,
  identifier: `https://${id}.example.com`,
});

describe('deriveResourceServerType', () => {
  it('reads a resource server with no interfaces as Custom', () => {
    expect(deriveResourceServerType([])).toBe('CUSTOM');
  });

  it('reads a resource server whose interfaces are not loaded as Custom', () => {
    expect(deriveResourceServerType(undefined)).toBe('CUSTOM');
  });

  it('reads an API-only resource server as API', () => {
    expect(deriveResourceServerType([iface('API', 'a')])).toBe('API');
  });

  it('reads an MCP resource server as MCP', () => {
    expect(deriveResourceServerType([iface('MCP', 'm')])).toBe('MCP');
  });

  // The badge is deliberately lossy: a resource server exposing both audiences reads as MCP, and the
  // Interfaces tab lists every interface for the exact picture.
  it('reads a resource server exposing both an API and an MCP interface as MCP', () => {
    expect(deriveResourceServerType([iface('API', 'a'), iface('MCP', 'm')])).toBe('MCP');
  });

  it('does not depend on interface order', () => {
    expect(deriveResourceServerType([iface('MCP', 'm'), iface('API', 'a')])).toBe('MCP');
  });
});
