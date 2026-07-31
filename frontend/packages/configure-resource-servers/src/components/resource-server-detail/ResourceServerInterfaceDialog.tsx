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

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Typography,
} from '@wso2/oxygen-ui';
import {useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {
  ResourceServerInterface,
  ResourceServerInterfaceRequest,
  ResourceServerInterfaceType,
} from '../../models/resource-server';

export interface ResourceServerInterfaceDialogProps {
  open: boolean;
  target: ResourceServerInterface | null;
  submitting: boolean;
  onClose: () => void;
  onSubmit: (data: ResourceServerInterfaceRequest) => void;
}

const INTERFACE_TYPES: ResourceServerInterfaceType[] = ['API', 'MCP'];

export default function ResourceServerInterfaceDialog({
  open,
  target,
  submitting,
  onClose,
  onSubmit,
}: ResourceServerInterfaceDialogProps): JSX.Element {
  const {t} = useTranslation();
  const [type, setType] = useState<ResourceServerInterfaceType>(target?.type ?? 'API');
  const [identifier, setIdentifier] = useState(target?.identifier ?? '');

  const trimmed = identifier.trim();
  const unchanged = target !== null && target.type === type && target.identifier === trimmed;
  const canSubmit = trimmed.length > 0 && !unchanged && !submitting;

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <DialogTitle>
        {target
          ? t('resourceServers:edit.interfaces.dialog.editTitle', 'Edit interface')
          : t('resourceServers:edit.interfaces.dialog.addTitle', 'Add interface')}
      </DialogTitle>
      <DialogContent>
        <Stack spacing={2.5} sx={{pt: 1}}>
          <Typography variant="body2" color="text.secondary">
            {t(
              'resourceServers:edit.interfaces.dialog.description',
              'The identifier is the audience of access tokens issued for this interface, and must be unique across the deployment.',
            )}
          </Typography>

          <FormControl fullWidth>
            <FormLabel htmlFor="resource-server-interface-type">
              {t('resourceServers:edit.interfaces.dialog.typeLabel', 'Type')}
            </FormLabel>
            <Select
              id="resource-server-interface-type"
              value={type}
              onChange={(e) => setType(e.target.value as ResourceServerInterfaceType)}
              data-testid="resource-server-interface-type-select"
            >
              {INTERFACE_TYPES.map((option) => (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          <FormControl fullWidth>
            <FormLabel htmlFor="resource-server-interface-identifier">
              {t('resourceServers:edit.interfaces.dialog.identifierLabel', 'Identifier (Audience)')}
            </FormLabel>
            <TextField
              id="resource-server-interface-identifier"
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              fullWidth
              size="small"
              placeholder={
                type === 'MCP'
                  ? t('resourceServers:edit.interfaces.dialog.placeholderMcp', 'https://mcp.example.com')
                  : t('resourceServers:edit.interfaces.dialog.placeholder', 'https://api.example.com')
              }
              helperText={t(
                'resourceServers:edit.interfaces.dialog.identifierHint',
                'Clients pass this value in the resource parameter to request tokens for this interface.',
              )}
              data-testid="resource-server-interface-identifier-input"
            />
          </FormControl>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={submitting}>
          {t('common:actions.cancel', 'Cancel')}
        </Button>
        <Button
          variant="contained"
          disabled={!canSubmit}
          onClick={() => onSubmit({type, identifier: trimmed})}
          data-testid="resource-server-interface-submit"
        >
          {target ? t('common:actions.save', 'Save') : t('common:actions.add', 'Add')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
