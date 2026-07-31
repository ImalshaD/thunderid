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
import type {
  DefaultResourceServerConfigResponse,
  ResourceServer,
  ResourceServerInterfaceListResponse,
} from '../../../models/resource-server';
import ResourceServerInterfacesSection from '../ResourceServerInterfacesSection';

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

const mockUseGetInterfaces = vi.fn();
const mockCreate = vi.fn();
const mockUpdate = vi.fn();
const mockDelete = vi.fn();
const mockUseGetDefault = vi.fn();

vi.mock('../../../api/useGetResourceServerInterfaces', () => ({
  default: () =>
    mockUseGetInterfaces() as {
      data: ResourceServerInterfaceListResponse | undefined;
      isLoading: boolean;
      error: Error | null;
    },
}));

vi.mock('../../../api/useCreateResourceServerInterface', () => ({
  default: () => ({mutate: mockCreate, isPending: false}),
}));

vi.mock('../../../api/useUpdateResourceServerInterface', () => ({
  default: () => ({mutate: mockUpdate, isPending: false}),
}));

vi.mock('../../../api/useDeleteResourceServerInterface', () => ({
  default: () => ({mutate: mockDelete, isPending: false}),
}));

vi.mock('../../../api/useGetDefaultResourceServer', () => ({
  default: () => mockUseGetDefault() as {data: DefaultResourceServerConfigResponse | undefined},
}));

const resourceServer: ResourceServer = {
  id: 'rs-1',
  name: 'Orders',
  ouId: 'ou-1',
  delimiter: ':',
  interfaces: [],
};

describe('ResourceServerInterfacesSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetDefault.mockReturnValue({
      data: {readOnly: {}, writable: {}, merged: {resourceServerInterfaceId: 'rsi-api'}},
    });
    mockUseGetInterfaces.mockReturnValue({
      data: {
        totalResults: 2,
        interfaces: [
          {id: 'rsi-api', type: 'API', identifier: 'https://api.example.com/orders'},
          {id: 'rsi-mcp', type: 'MCP', identifier: 'https://api.example.com/orders/mcp'},
        ],
      },
      isLoading: false,
      error: null,
    });
  });

  it('lists each interface with its type and identifier', () => {
    renderWithProviders(<ResourceServerInterfacesSection resourceServer={resourceServer} />);

    expect(screen.getByText('https://api.example.com/orders')).toBeInTheDocument();
    expect(screen.getByText('https://api.example.com/orders/mcp')).toBeInTheDocument();
  });

  it('marks the interface that is the deployment default', () => {
    renderWithProviders(<ResourceServerInterfacesSection resourceServer={resourceServer} />);

    expect(screen.getByText('Default')).toBeInTheDocument();
  });

  // The backend refuses while the interface holds the default, so the reason is surfaced up front.
  it('blocks deleting the default interface and allows deleting the other', () => {
    renderWithProviders(<ResourceServerInterfacesSection resourceServer={resourceServer} />);

    expect(screen.getByTestId('resource-server-interface-delete-rsi-api')).toBeDisabled();
    expect(screen.getByTestId('resource-server-interface-delete-rsi-mcp')).toBeEnabled();
  });

  it('adds an interface through the dialog', async () => {
    renderWithProviders(<ResourceServerInterfacesSection resourceServer={resourceServer} />);

    await userEvent.click(screen.getByTestId('resource-server-interface-add'));
    await userEvent.type(
      screen.getByTestId('resource-server-interface-identifier-input').querySelector('input')!,
      'https://api.example.com/v2',
    );
    await userEvent.click(screen.getByTestId('resource-server-interface-submit'));

    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith(
        {type: 'API', identifier: 'https://api.example.com/v2'},
        expect.anything(),
      );
    });
  });

  it('edits an interface through the same dialog', async () => {
    renderWithProviders(<ResourceServerInterfacesSection resourceServer={resourceServer} />);

    await userEvent.click(screen.getByTestId('resource-server-interface-edit-rsi-mcp'));
    const input = screen.getByTestId('resource-server-interface-identifier-input').querySelector('input')!;
    await userEvent.clear(input);
    await userEvent.type(input, 'https://api.example.com/orders/mcp/v2');
    await userEvent.click(screen.getByTestId('resource-server-interface-submit'));

    await waitFor(() => {
      expect(mockUpdate).toHaveBeenCalledWith(
        {interfaceId: 'rsi-mcp', data: {type: 'MCP', identifier: 'https://api.example.com/orders/mcp/v2'}},
        expect.anything(),
      );
    });
  });

  it('deletes an interface after confirmation', async () => {
    renderWithProviders(<ResourceServerInterfacesSection resourceServer={resourceServer} />);

    await userEvent.click(screen.getByTestId('resource-server-interface-delete-rsi-mcp'));
    await userEvent.click(screen.getByTestId('resource-server-interface-delete-confirm'));

    await waitFor(() => {
      expect(mockDelete).toHaveBeenCalledWith('rsi-mcp', expect.anything());
    });
  });

  // A resource server with no interfaces is legitimate: it defines permissions but has no audience.
  it('explains the empty state', () => {
    mockUseGetInterfaces.mockReturnValue({
      data: {totalResults: 0, interfaces: []},
      isLoading: false,
      error: null,
    });

    renderWithProviders(<ResourceServerInterfacesSection resourceServer={resourceServer} />);

    expect(screen.getByTestId('resource-server-interfaces-empty')).toBeInTheDocument();
  });

  it('hides every action for a read-only resource server', () => {
    renderWithProviders(
      <ResourceServerInterfacesSection resourceServer={{...resourceServer, isReadOnly: true}} />,
    );

    expect(screen.queryByTestId('resource-server-interface-add')).not.toBeInTheDocument();
    expect(screen.queryByTestId('resource-server-interface-edit-rsi-mcp')).not.toBeInTheDocument();
  });
});
