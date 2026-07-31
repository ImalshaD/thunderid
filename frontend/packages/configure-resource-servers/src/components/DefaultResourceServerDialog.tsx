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

import {useToast} from '@thunderid/contexts';
import {useLogger} from '@thunderid/logger/react';
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  FormLabel,
  MenuItem,
  Radio,
  RadioGroup,
  Select,
  Stack,
  Typography,
} from '@wso2/oxygen-ui';
import {useMemo, useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useSetDefaultResourceServer from '../api/useSetDefaultResourceServer';
import type {ResourceServer} from '../models/resource-server';

export interface DefaultResourceServerDialogProps {
  resourceServers: ResourceServer[];
  currentInterfaceId: string | undefined;
  onClose: () => void;
}

export default function DefaultResourceServerDialog({
  resourceServers,
  currentInterfaceId,
  onClose,
}: DefaultResourceServerDialogProps): JSX.Element {
  const {t} = useTranslation();
  const {showToast} = useToast();
  const logger = useLogger('DefaultResourceServerDialog');
  const setDefault = useSetDefaultResourceServer();

  // A resource server with no interfaces cannot supply an audience, so it is not offered at all
  // rather than shown and disabled.
  const selectable = useMemo(
    () => resourceServers.filter((rs) => (rs.interfaces?.length ?? 0) > 0),
    [resourceServers],
  );

  const owningServer = selectable.find((rs) => rs.interfaces.some((rsi) => rsi.id === currentInterfaceId));
  const initialServer = owningServer ?? selectable[0];

  const [serverId, setServerId] = useState(initialServer?.id ?? '');
  const [interfaceId, setInterfaceId] = useState(
    owningServer && currentInterfaceId ? currentInterfaceId : (initialServer?.interfaces[0]?.id ?? ''),
  );
  const selectedServer = selectable.find((rs) => rs.id === serverId);

  const handleServerChange = (nextServerId: string): void => {
    setServerId(nextServerId);
    const next = selectable.find((rs) => rs.id === nextServerId);
    setInterfaceId(next?.interfaces[0]?.id ?? '');
  };

  const submit = (nextInterfaceId: string, successMessage: string): void => {
    setDefault.mutate(
      {resourceServerInterfaceId: nextInterfaceId},
      {
        onSuccess: () => {
          showToast(successMessage, 'success');
          onClose();
        },
        onError: (err: Error) => {
          logger.error('Failed to update the default resource server interface', {error: err});
          showToast(
            t('resourceServers:default.saveError', 'Failed to update the default resource server.'),
            'error',
          );
        },
      },
    );
  };

  const noneSelectable = selectable.length === 0;

  return (
    <Dialog open onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>{t('resourceServers:default.dialog.title', 'Default resource server')}</DialogTitle>
      <DialogContent>
        <Stack spacing={2.5} sx={{pt: 1}}>
          <Typography variant="body2" color="text.secondary">
            {t(
              'resourceServers:default.dialog.description',
              'Access tokens requested without a resource parameter are issued for the interface you select, and carry its identifier as the audience.',
            )}
          </Typography>

          {noneSelectable && (
            <Alert severity="info" data-testid="default-resource-server-no-options">
              {t(
                'resourceServers:default.dialog.noneSelectable',
                'No resource server has an interface yet. Add an interface to a resource server before setting a default.',
              )}
            </Alert>
          )}

          {!noneSelectable && (
            <>
              <FormControl fullWidth>
                <FormLabel htmlFor="default-resource-server-select">
                  {t('resourceServers:default.dialog.resourceServerLabel', 'Resource server')}
                </FormLabel>
                <Select
                  id="default-resource-server-select"
                  value={serverId}
                  onChange={(e) => handleServerChange(e.target.value)}
                  data-testid="default-resource-server-select"
                >
                  {selectable.map((rs) => (
                    <MenuItem key={rs.id} value={rs.id}>
                      {rs.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              <FormControl>
                <FormLabel id="default-resource-server-interface-label">
                  {t('resourceServers:default.dialog.interfaceLabel', 'Interface')}
                </FormLabel>
                <RadioGroup
                  aria-labelledby="default-resource-server-interface-label"
                  value={interfaceId}
                  onChange={(e) => setInterfaceId(e.target.value)}
                >
                  {selectedServer?.interfaces.map((rsi) => (
                    <FormControlLabel
                      key={rsi.id}
                      value={rsi.id}
                      control={<Radio size="small" />}
                      data-testid={`default-resource-server-interface-${rsi.id}`}
                      label={
                        <Box sx={{display: 'flex', alignItems: 'baseline', gap: 1}}>
                          <Typography variant="body2" fontWeight={500}>
                            {rsi.type}
                          </Typography>
                          <Typography
                            variant="body2"
                            color="text.secondary"
                            sx={{fontFamily: 'monospace', fontSize: '0.8rem', overflowWrap: 'anywhere'}}
                          >
                            {rsi.identifier}
                          </Typography>
                        </Box>
                      }
                    />
                  ))}
                </RadioGroup>
              </FormControl>
            </>
          )}
        </Stack>
      </DialogContent>
      <DialogActions sx={{justifyContent: 'space-between'}}>
        <Button
          color="error"
          disabled={!currentInterfaceId || setDefault.isPending}
          onClick={() => submit('', t('resourceServers:default.cleared', 'Default resource server cleared.'))}
          data-testid="default-resource-server-clear"
        >
          {t('resourceServers:default.dialog.clear', 'Clear default')}
        </Button>
        <Stack direction="row" spacing={1}>
          <Button onClick={onClose} disabled={setDefault.isPending}>
            {t('common:actions.cancel', 'Cancel')}
          </Button>
          <Button
            variant="contained"
            disabled={!interfaceId || setDefault.isPending}
            onClick={() => submit(interfaceId, t('resourceServers:default.saved', 'Default resource server updated.'))}
            data-testid="default-resource-server-save"
          >
            {t('common:actions.save', 'Save')}
          </Button>
        </Stack>
      </DialogActions>
    </Dialog>
  );
}
