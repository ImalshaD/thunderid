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

import * as thunderIdReactModule from '@thunderid/react';
import {renderWithProviders, screen} from '@thunderid/test-utils';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import type {DefaultResourceServerConfigResponse, ResourceServer} from '../../models/resource-server';
import DefaultResourceServerCard from '../DefaultResourceServerCard';

vi.mock('@thunderid/react', {spy: true});

// eslint-disable-next-line @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-member-access -- vi.mock({spy:true}) type inference doesn't resolve for this package's conditional exports
vi.mocked(thunderIdReactModule.useThunderID).mockImplementation(() => ({http: {request: vi.fn()}}) as never);

vi.mock('@thunderid/contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/contexts')>();
  return {
    ...actual,
    useConfig: () => ({getServerUrl: () => 'http://localhost:8090'}),
    useToast: () => ({showToast: vi.fn()}),
  };
});

vi.mock('@thunderid/logger/react', () => ({
  useLogger: () => ({error: vi.fn(), info: vi.fn(), debug: vi.fn()}),
}));

vi.mock('../DefaultResourceServerDialog', () => ({
  default: () => null,
}));

const mockUseGetDefaultResourceServer = vi.fn();

vi.mock('../../api/useGetDefaultResourceServer', () => ({
  default: () => mockUseGetDefaultResourceServer() as {data: DefaultResourceServerConfigResponse | undefined},
}));

const resourceServers: ResourceServer[] = [
  {
    id: 'rs-1',
    name: 'Orders',
    ouId: 'ou-1',
    delimiter: ':',
    interfaces: [
      {id: 'rsi-api', type: 'API', identifier: 'https://api.example.com/orders'},
      {id: 'rsi-mcp', type: 'MCP', identifier: 'https://api.example.com/orders/mcp'},
    ],
  },
];

describe('DefaultResourceServerCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows the owning resource server and the configured interface identifier', () => {
    mockUseGetDefaultResourceServer.mockReturnValue({
      data: {readOnly: {}, writable: {}, merged: {resourceServerInterfaceId: 'rsi-mcp'}},
    });

    renderWithProviders(<DefaultResourceServerCard resourceServers={resourceServers} />);

    expect(screen.getByText('Orders')).toBeInTheDocument();
    // The identifier is the audience tokens carry, so it is shown rather than only the name.
    expect(screen.getByText('https://api.example.com/orders/mcp')).toBeInTheDocument();
    expect(screen.getByText('MCP')).toBeInTheDocument();
  });

  it('explains the consequence when no default is configured', () => {
    mockUseGetDefaultResourceServer.mockReturnValue({
      data: {readOnly: {}, writable: {}, merged: {}},
    });

    renderWithProviders(<DefaultResourceServerCard resourceServers={resourceServers} />);

    expect(screen.getByTestId('default-resource-server-empty')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Select'})).toBeEnabled();
  });

  // A declarative default cannot be written at runtime, so the action is disabled rather than failing.
  it('disables editing a declaratively configured default', () => {
    mockUseGetDefaultResourceServer.mockReturnValue({
      data: {
        readOnly: {resourceServerInterfaceId: 'rsi-api'},
        writable: {},
        merged: {resourceServerInterfaceId: 'rsi-api'},
      },
    });

    renderWithProviders(<DefaultResourceServerCard resourceServers={resourceServers} />);

    expect(screen.getByTestId('default-resource-server-edit')).toBeDisabled();
    expect(screen.getByText('Declarative')).toBeInTheDocument();
  });

  // The listing is paginated, so the owning resource server may not be on the current page.
  it('falls back to the interface ID when the owning resource server is not loaded', () => {
    mockUseGetDefaultResourceServer.mockReturnValue({
      data: {readOnly: {}, writable: {}, merged: {resourceServerInterfaceId: 'rsi-elsewhere'}},
    });

    renderWithProviders(<DefaultResourceServerCard resourceServers={resourceServers} />);

    expect(screen.getByText('rsi-elsewhere')).toBeInTheDocument();
  });
});
