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

import {SettingsCard} from '@thunderid/components';
import {Box, Button, Chip, Stack, Tooltip, Typography} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import DefaultResourceServerDialog from './DefaultResourceServerDialog';
import useGetDefaultResourceServer from '../api/useGetDefaultResourceServer';
import type {ResourceServer, ResourceServerInterface} from '../models/resource-server';

export interface DefaultResourceServerCardProps {
  resourceServers: ResourceServer[];
}

function resolveConfiguredInterface(
  resourceServers: ResourceServer[],
  configuredInterfaceId: string | undefined,
): {server: ResourceServer; rsInterface: ResourceServerInterface} | undefined {
  if (!configuredInterfaceId) return undefined;
  for (const server of resourceServers) {
    const rsInterface = server.interfaces?.find((rsi) => rsi.id === configuredInterfaceId);
    if (rsInterface) return {server, rsInterface};
  }
  return undefined;
}

export default function DefaultResourceServerCard({
  resourceServers,
}: DefaultResourceServerCardProps): JSX.Element {
  const {t} = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);

  const {data: config} = useGetDefaultResourceServer();
  const configuredInterfaceId = config?.merged?.resourceServerInterfaceId;
  // A default set through declarative configuration cannot be changed at runtime: the backend rejects
  // the write, so the action is disabled instead of failing.
  const isDeclarative = Boolean(config?.readOnly?.resourceServerInterfaceId);

  const resolved: {server: ResourceServer; rsInterface: ResourceServerInterface} | undefined =
    resolveConfiguredInterface(resourceServers, configuredInterfaceId);

  return (
    <>
      <SettingsCard
        title={t('resourceServers:default.title', 'Default resource server')}
        description={t(
          'resourceServers:default.description',
          'Access tokens requested without a resource parameter are issued for this interface.',
        )}
        headerAction={
          <Tooltip
            title={
              isDeclarative
                ? t(
                    'resourceServers:default.declarativeHint',
                    'Configured through declarative configuration and cannot be changed here.',
                  )
                : ''
            }
          >
            <span>
              <Button
                variant="outlined"
                size="small"
                disabled={isDeclarative}
                onClick={() => setDialogOpen(true)}
                data-testid="default-resource-server-edit"
              >
                {configuredInterfaceId
                  ? t('common:actions.edit', 'Edit')
                  : t('resourceServers:default.select', 'Select')}
              </Button>
            </span>
          </Tooltip>
        }
      >
        {configuredInterfaceId ? (
          <Stack direction="column" spacing={0.5} data-testid="default-resource-server-value">
            <Stack direction="row" spacing={1} alignItems="center">
              <Typography variant="body2" fontWeight={500}>
                {resolved?.server.name ??
                  t('resourceServers:default.unresolvedName', 'Resource server outside the current page')}
              </Typography>
              {isDeclarative && (
                <Chip
                  label={t('resourceServers:default.declarative', 'Declarative')}
                  size="small"
                  variant="outlined"
                  sx={{height: 20, fontSize: '0.65rem'}}
                />
              )}
            </Stack>
            <Box sx={{display: 'flex', alignItems: 'baseline', gap: 1}}>
              {resolved && (
                <Typography variant="body2" color="text.secondary">
                  {resolved.rsInterface.type}
                </Typography>
              )}
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{fontFamily: 'monospace', fontSize: '0.8rem', overflowWrap: 'anywhere'}}
              >
                {resolved?.rsInterface.identifier ?? configuredInterfaceId}
              </Typography>
            </Box>
          </Stack>
        ) : (
          <Typography variant="body2" color="text.secondary" data-testid="default-resource-server-empty">
            {t(
              'resourceServers:default.empty',
              'No default configured. Requests that ask for permissions without a resource parameter are rejected.',
            )}
          </Typography>
        )}
      </SettingsCard>

      {/* Mounted only while open, and keyed on the configured value, so it always opens on the
          current selection without an effect to reset its state. */}
      {dialogOpen && (
        <DefaultResourceServerDialog
          key={configuredInterfaceId ?? 'none'}
          resourceServers={resourceServers}
          currentInterfaceId={configuredInterfaceId}
          onClose={() => setDialogOpen(false)}
        />
      )}
    </>
  );
}
