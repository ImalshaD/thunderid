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
import {renderWithProviders, screen, userEvent, waitFor} from '@thunderid/test-utils';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import type {ResourceServer} from '../../models/resource-server';
import DefaultResourceServerDialog from '../DefaultResourceServerDialog';

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

const mockMutate = vi.fn();

vi.mock('../../api/useSetDefaultResourceServer', () => ({
  default: () => ({mutate: mockMutate, isPending: false}),
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
  {
    id: 'rs-2',
    name: 'Invoices',
    ouId: 'ou-1',
    delimiter: ':',
    // No interfaces: cannot supply an audience, so it must not be offered.
    interfaces: [],
  },
];

describe('DefaultResourceServerDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('preselects the configured interface so it doubles as changing which interface is default', () => {
    renderWithProviders(
      <DefaultResourceServerDialog
        resourceServers={resourceServers}
        currentInterfaceId="rsi-mcp"
        onClose={vi.fn()}
      />,
    );

    expect(screen.getByTestId('default-resource-server-interface-rsi-mcp')).toBeInTheDocument();
    expect(screen.getByRole('radio', {name: /orders\/mcp/})).toBeChecked();
  });

  it('saves the selected interface ID', async () => {
    renderWithProviders(
      <DefaultResourceServerDialog
        resourceServers={resourceServers}
        currentInterfaceId="rsi-mcp"
        onClose={vi.fn()}
      />,
    );

    await userEvent.click(screen.getByRole('radio', {name: /orders$/}));
    await userEvent.click(screen.getByTestId('default-resource-server-save'));

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith({resourceServerInterfaceId: 'rsi-api'}, expect.anything());
    });
  });

  it('omits resource servers that have no interfaces', () => {
    renderWithProviders(
      <DefaultResourceServerDialog
        resourceServers={resourceServers}
        currentInterfaceId={undefined}
        onClose={vi.fn()}
      />,
    );

    expect(screen.queryByText('Invoices')).not.toBeInTheDocument();
  });

  it('explains that nothing can be selected when no resource server has an interface', () => {
    renderWithProviders(
      <DefaultResourceServerDialog
        resourceServers={resourceServers.slice(1)}
        currentInterfaceId={undefined}
        onClose={vi.fn()}
      />,
    );

    expect(screen.getByTestId('default-resource-server-no-options')).toBeInTheDocument();
    expect(screen.getByTestId('default-resource-server-save')).toBeDisabled();
  });

  it('clears the default with an empty interface ID', async () => {
    renderWithProviders(
      <DefaultResourceServerDialog
        resourceServers={resourceServers}
        currentInterfaceId="rsi-api"
        onClose={vi.fn()}
      />,
    );

    await userEvent.click(screen.getByTestId('default-resource-server-clear'));

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith({resourceServerInterfaceId: ''}, expect.anything());
    });
  });
});
