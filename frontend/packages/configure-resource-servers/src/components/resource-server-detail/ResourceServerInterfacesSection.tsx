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
import {useToast} from '@thunderid/contexts';
import {useLogger} from '@thunderid/logger/react';
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui';
import {Pencil, Plus, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import ResourceServerInterfaceDialog from './ResourceServerInterfaceDialog';
import useCreateResourceServerInterface from '../../api/useCreateResourceServerInterface';
import useDeleteResourceServerInterface from '../../api/useDeleteResourceServerInterface';
import useGetDefaultResourceServer from '../../api/useGetDefaultResourceServer';
import useGetResourceServerInterfaces from '../../api/useGetResourceServerInterfaces';
import useUpdateResourceServerInterface from '../../api/useUpdateResourceServerInterface';
import type {
  ResourceServer,
  ResourceServerInterface,
  ResourceServerInterfaceRequest,
} from '../../models/resource-server';

export interface ResourceServerInterfacesSectionProps {
  resourceServer: ResourceServer;
}

export default function ResourceServerInterfacesSection({
  resourceServer,
}: ResourceServerInterfacesSectionProps): JSX.Element {
  const {t} = useTranslation();
  const {showToast} = useToast();
  const logger = useLogger('ResourceServerInterfacesSection');

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<ResourceServerInterface | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ResourceServerInterface | null>(null);

  const {data, isLoading, error} = useGetResourceServerInterfaces(resourceServer.id);
  const {data: defaultConfig} = useGetDefaultResourceServer();
  const defaultInterfaceId = defaultConfig?.merged?.resourceServerInterfaceId;

  const createInterface = useCreateResourceServerInterface(resourceServer.id);
  const updateInterface = useUpdateResourceServerInterface(resourceServer.id);
  const deleteInterface = useDeleteResourceServerInterface(resourceServer.id);

  const interfaces = data?.interfaces ?? [];
  const isReadOnly = Boolean(resourceServer.isReadOnly);
  const submitting = createInterface.isPending || updateInterface.isPending;

  const closeDialog = (): void => {
    setDialogOpen(false);
    setEditTarget(null);
  };

  const handleSubmit = (values: ResourceServerInterfaceRequest): void => {
    const onSuccess = (): void => {
      showToast(
        editTarget
          ? t('resourceServers:edit.interfaces.updated', 'Interface updated.')
          : t('resourceServers:edit.interfaces.added', 'Interface added.'),
        'success',
      );
      closeDialog();
    };
    const onError = (err: Error): void => {
      logger.error('Failed to save resource server interface', {error: err});
      showToast(t('resourceServers:edit.interfaces.saveError', 'Failed to save the interface.'), 'error');
    };

    if (editTarget) {
      updateInterface.mutate({interfaceId: editTarget.id, data: values}, {onSuccess, onError});
      return;
    }
    createInterface.mutate(values, {onSuccess, onError});
  };

  const handleDelete = (): void => {
    if (!deleteTarget) return;
    deleteInterface.mutate(deleteTarget.id, {
      onSuccess: () => {
        showToast(t('resourceServers:edit.interfaces.deleted', 'Interface deleted.'), 'success');
        setDeleteTarget(null);
      },
      onError: (err: Error) => {
        logger.error('Failed to delete resource server interface', {error: err});
        showToast(t('resourceServers:edit.interfaces.deleteError', 'Failed to delete the interface.'), 'error');
        setDeleteTarget(null);
      },
    });
  };

  return (
    <SettingsCard
      title={t('resourceServers:edit.interfaces.title', 'Interfaces')}
      description={t(
        'resourceServers:edit.interfaces.description',
        'The ways this resource server is accessed. Each interface has its own audience identifier, while permissions and role assignments stay on the resource server and are shared by all of them.',
      )}
      headerAction={
        isReadOnly ? undefined : (
          <Button
            variant="outlined"
            size="small"
            startIcon={<Plus size={16} />}
            onClick={() => {
              setEditTarget(null);
              setDialogOpen(true);
            }}
            data-testid="resource-server-interface-add"
          >
            {t('resourceServers:edit.interfaces.add', 'Add interface')}
          </Button>
        )
      }
    >
      <Stack spacing={2}>
        {error && (
          <Alert severity="error">
            {t('resourceServers:edit.interfaces.loadError', 'Failed to load interfaces.')}
          </Alert>
        )}

        {isLoading && <CircularProgress size={20} />}

        {!isLoading && interfaces.length === 0 && (
          <Typography variant="body2" color="text.secondary" data-testid="resource-server-interfaces-empty">
            {t(
              'resourceServers:edit.interfaces.empty',
              'No interfaces yet. This resource server defines permissions but cannot receive tokens until an interface is added.',
            )}
          </Typography>
        )}

        {interfaces.map((rsi) => {
          const isDefault = rsi.id === defaultInterfaceId;
          const rowReadOnly = isReadOnly || Boolean(rsi.isReadOnly);

          return (
            <Box
              key={rsi.id}
              data-testid={`resource-server-interface-row-${rsi.id}`}
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 2,
                px: 2,
                py: 1.5,
                border: 1,
                borderColor: 'divider',
                borderRadius: 1,
              }}
            >
              <Stack direction="row" spacing={1.5} alignItems="center" sx={{minWidth: 0}}>
                <Chip label={rsi.type} size="small" color="primary" variant="outlined" sx={{fontSize: '0.7rem'}} />
                <Typography
                  variant="body2"
                  sx={{fontFamily: 'monospace', fontSize: '0.8rem', overflowWrap: 'anywhere'}}
                >
                  {rsi.identifier}
                </Typography>
                {isDefault && (
                  <Chip
                    label={t('resourceServers:edit.interfaces.default', 'Default')}
                    size="small"
                    color="success"
                    variant="outlined"
                    sx={{height: 20, fontSize: '0.65rem'}}
                  />
                )}
              </Stack>

              {!rowReadOnly && (
                <Stack direction="row" spacing={0.5}>
                  <Tooltip title={t('common:actions.edit', 'Edit')}>
                    <IconButton
                      size="small"
                      onClick={() => {
                        setEditTarget(rsi);
                        setDialogOpen(true);
                      }}
                      data-testid={`resource-server-interface-edit-${rsi.id}`}
                    >
                      <Pencil size={16} />
                    </IconButton>
                  </Tooltip>
                  {/* The deployment default cannot be deleted: the backend refuses while it holds that
                      role, so surface the reason rather than a failed request. */}
                  <Tooltip
                    title={
                      isDefault
                        ? t(
                            'resourceServers:edit.interfaces.deleteDefaultBlocked',
                            'This interface is the default resource server. Change the default before deleting it.',
                          )
                        : t('common:actions.delete', 'Delete')
                    }
                  >
                    <span>
                      <IconButton
                        size="small"
                        color="error"
                        disabled={isDefault}
                        onClick={() => setDeleteTarget(rsi)}
                        data-testid={`resource-server-interface-delete-${rsi.id}`}
                      >
                        <Trash2 size={16} />
                      </IconButton>
                    </span>
                  </Tooltip>
                </Stack>
              )}
            </Box>
          );
        })}
      </Stack>

      {dialogOpen && (
        <ResourceServerInterfaceDialog
          key={editTarget?.id ?? 'new'}
          open
          target={editTarget}
          submitting={submitting}
          onClose={closeDialog}
          onSubmit={handleSubmit}
        />
      )}

      <Dialog open={deleteTarget !== null} onClose={() => setDeleteTarget(null)} maxWidth="xs" fullWidth>
        <DialogTitle>{t('resourceServers:edit.interfaces.deleteTitle', 'Delete interface')}</DialogTitle>
        <DialogContent>
          <Alert severity="warning" sx={{mb: 2}}>
            {t(
              'resourceServers:edit.interfaces.deleteWarning',
              'Clients can no longer request tokens for this audience.',
            )}
          </Alert>
          <Typography variant="body2" color="text.secondary">
            {t(
              'resourceServers:edit.interfaces.deleteDescription',
              'Tokens already issued for this audience stay valid until they expire.',
            )}
          </Typography>
          {deleteTarget && (
            <Typography variant="body2" sx={{fontFamily: 'monospace', fontSize: '0.8rem', mt: 1.5}}>
              {deleteTarget.identifier}
            </Typography>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)} disabled={deleteInterface.isPending}>
            {t('common:actions.cancel', 'Cancel')}
          </Button>
          <Button
            variant="contained"
            color="error"
            onClick={handleDelete}
            disabled={deleteInterface.isPending}
            data-testid="resource-server-interface-delete-confirm"
          >
            {t('common:actions.delete', 'Delete')}
          </Button>
        </DialogActions>
      </Dialog>
    </SettingsCard>
  );
}
