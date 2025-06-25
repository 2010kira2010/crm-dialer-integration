import React, { memo, useState, useEffect } from 'react';
import { Handle, Position } from 'react-flow-renderer';
import {
    Box,
    Paper,
    Typography,
    Select,
    MenuItem,
    FormControl,
    InputLabel,
    TextField,
    Chip,
    Divider,
} from '@mui/material';
import CallIcon from '@mui/icons-material/Call';
import UpdateIcon from '@mui/icons-material/Update';
import PriorityHighIcon from '@mui/icons-material/PriorityHigh';
import ScheduleIcon from '@mui/icons-material/Schedule';
import RemoveCircleIcon from '@mui/icons-material/RemoveCircle';
import { useStores } from '../../../hooks/useStores';
import { observer } from 'mobx-react-lite';
import { ActionType } from '../../../types';

const actionIcons: Record<ActionType, React.ReactNode> = {
    update_lead: <UpdateIcon />,
    add_to_bucket: <CallIcon />,
    change_priority: <PriorityHighIcon />,
    change_scheduler_step: <ScheduleIcon />,
    remove_from_dialer: <RemoveCircleIcon />,
};

const actionLabels: Record<ActionType, string> = {
    update_lead: 'Обновить лид',
    add_to_bucket: 'Добавить в бакет',
    change_priority: 'Изменить приоритет',
    change_scheduler_step: 'Изменить шаг шедуллера',
    remove_from_dialer: 'Изъять из прозвона',
};

interface ActionNodeProps {
    data: any;
    id: string;
    selected?: boolean;
}

const ActionNodeComponent: React.FC<ActionNodeProps> = ({ data, id, selected }) => {
    const { dataStore, flowStore } = useStores();
    const [actionType, setActionType] = useState<ActionType>(data.actionType || 'update_lead');
    const [actionData, setActionData] = useState<any>(data.actionData || {});

    useEffect(() => {
        if (dataStore.amocrmFields.length === 0) {
            dataStore.loadAllData();
        }
    }, [dataStore]);

    useEffect(() => {
        flowStore.updateNodeData(id, { actionType, actionData });
    }, [actionType, actionData, id, flowStore]);

    const handleActionTypeChange = (type: ActionType) => {
        setActionType(type);
        // Reset action data when changing type
        switch (type) {
            case 'update_lead':
                setActionData({ fields: {}, status_id: '', pipeline_id: '' });
                break;
            case 'add_to_bucket':
                setActionData({ bucket_id: '', priority: 50, scheduler_id: '', scheduler_step: 1 });
                break;
            case 'change_priority':
                setActionData({ priority: 50 });
                break;
            case 'change_scheduler_step':
                setActionData({ scheduler_step: 1 });
                break;
            case 'remove_from_dialer':
                setActionData({});
                break;
        }
    };

    const updateActionData = (key: string, value: any) => {
        setActionData((prev: any) => ({ ...prev, [key]: value }));
    };

    const updateFieldValue = (fieldId: string, value: any) => {
        setActionData((prev: any) => ({
            ...prev,
            fields: {
                ...prev.fields,
                [fieldId]: value,
            },
        }));
    };

    const renderActionConfig = () => {
        switch (actionType) {
            case 'update_lead':
                return (
                    <>
                        <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                            Выберите поля для обновления:
                        </Typography>

                        <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                            <InputLabel>Воронка</InputLabel>
                            <Select
                                value={actionData.pipeline_id || ''}
                                onChange={(e) => {
                                    updateActionData('pipeline_id', e.target.value);
                                    // Reset status when pipeline changes
                                    updateActionData('status_id', '');
                                }}
                                label="Воронка"
                            >
                                <MenuItem value="">
                                    <em>Не изменять</em>
                                </MenuItem>
                                {dataStore.amocrmPipelines.map((pipeline) => (
                                    <MenuItem key={pipeline.id} value={pipeline.id}>
                                        {pipeline.name}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                            <InputLabel>Статус/Этап</InputLabel>
                            <Select
                                value={actionData.status_id || ''}
                                onChange={(e) => updateActionData('status_id', e.target.value)}
                                label="Статус/Этап"
                                disabled={!actionData.pipeline_id}
                            >
                                <MenuItem value="">
                                    <em>Не изменять</em>
                                </MenuItem>
                                {actionData.pipeline_id && dataStore.amocrmPipelines
                                    .find(p => p.id === actionData.pipeline_id)
                                    ?.statuses.map((status) => (
                                        <MenuItem key={status.id} value={status.id}>
                                            {status.name}
                                        </MenuItem>
                                    ))}
                            </Select>
                        </FormControl>

                        <Divider sx={{ my: 1 }} />

                        <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                            Кастомные поля:
                        </Typography>
                        {dataStore.amocrmFields.slice(0, 5).map((field) => (
                            <TextField
                                key={field.id}
                                fullWidth
                                size="small"
                                label={field.name}
                                value={actionData.fields?.[field.id] || ''}
                                onChange={(e) => updateFieldValue(field.id.toString(), e.target.value)}
                                sx={{ mb: 1 }}
                                helperText={`Тип: ${field.type}`}
                            />
                        ))}
                        {dataStore.amocrmFields.length > 5 && (
                            <Typography variant="caption" color="text.secondary">
                                + еще {dataStore.amocrmFields.length - 5} полей...
                            </Typography>
                        )}
                    </>
                );

            case 'add_to_bucket':
                return (
                    <>
                        <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                            <InputLabel>Бакет</InputLabel>
                            <Select
                                value={actionData.bucket_id || ''}
                                onChange={(e) => updateActionData('bucket_id', e.target.value)}
                                label="Бакет"
                                required
                            >
                                <MenuItem value="">
                                    <em>Выберите бакет</em>
                                </MenuItem>
                                {dataStore.dialerBuckets.map((bucket) => (
                                    <MenuItem key={bucket.id} value={bucket.id}>
                                        {bucket.name}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <TextField
                            fullWidth
                            size="small"
                            label="Приоритет в бакете"
                            type="number"
                            value={actionData.priority || 50}
                            onChange={(e) => updateActionData('priority', parseInt(e.target.value) || 0)}
                            sx={{ mb: 1 }}
                            inputProps={{ min: 0, max: 100 }}
                            helperText="0-100, чем выше число, тем выше приоритет"
                        />

                        <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                            <InputLabel>Шедуллер</InputLabel>
                            <Select
                                value={actionData.scheduler_id || ''}
                                onChange={(e) => updateActionData('scheduler_id', e.target.value)}
                                label="Шедуллер"
                                required
                            >
                                <MenuItem value="">
                                    <em>Выберите шедуллер</em>
                                </MenuItem>
                                {dataStore.dialerSchedulers.map((scheduler) => (
                                    <MenuItem key={scheduler.id} value={scheduler.id}>
                                        {scheduler.name}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <TextField
                            fullWidth
                            size="small"
                            label="Шаг шедуллера"
                            type="number"
                            value={actionData.scheduler_step || 1}
                            onChange={(e) => updateActionData('scheduler_step', parseInt(e.target.value) || 1)}
                            inputProps={{ min: 1 }}
                            helperText="Начальный шаг в шедуллере (обычно 1)"
                        />
                    </>
                );

            case 'change_priority':
                return (
                    <>
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                            Изменение приоритета для контакта, который уже находится в системе автообзвона
                        </Typography>
                        <TextField
                            fullWidth
                            size="small"
                            label="Новый приоритет"
                            type="number"
                            value={actionData.priority || 50}
                            onChange={(e) => updateActionData('priority', parseInt(e.target.value) || 0)}
                            inputProps={{ min: 0, max: 100 }}
                            helperText="0-100, чем выше число, тем выше приоритет"
                        />
                    </>
                );

            case 'change_scheduler_step':
                return (
                    <>
                        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                            Изменение текущего шага в шедуллере для контакта, который уже находится в системе автообзвона
                        </Typography>
                        <TextField
                            fullWidth
                            size="small"
                            label="Новый шаг шедуллера"
                            type="number"
                            value={actionData.scheduler_step || 1}
                            onChange={(e) => updateActionData('scheduler_step', parseInt(e.target.value) || 1)}
                            inputProps={{ min: 1 }}
                            helperText="Установить конкретный шаг в шедуллере"
                        />
                    </>
                );

            case 'remove_from_dialer':
                return (
                    <Box>
                        <Typography variant="body2" color="text.secondary">
                            Контакт будет полностью удален из системы автообзвона
                        </Typography>
                        <Typography variant="caption" color="error" sx={{ mt: 1, display: 'block' }}>
                            ⚠️ Это действие необратимо. Контакт будет удален из всех бакетов и кампаний.
                        </Typography>
                    </Box>
                );

            default:
                return null;
        }
    };

    return (
        <Paper
            sx={{
                p: 2,
                minWidth: 350,
                maxWidth: 450,
                border: selected ? '2px solid #1976d2' : '1px solid #ccc',
                bgcolor: selected ? 'action.hover' : 'background.paper',
                boxShadow: selected ? 3 : 1,
                transition: 'all 0.2s ease',
            }}
        >
            <Handle
                type="target"
                position={Position.Top}
                style={{ background: '#555' }}
            />

            <Box sx={{ mb: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                    {actionIcons[actionType]}
                    <Typography variant="subtitle2" fontWeight="bold">
                        Действие
                    </Typography>
                    {selected && (
                        <Chip label="Выбрано" size="small" color="primary" sx={{ ml: 'auto' }} />
                    )}
                </Box>

                <FormControl fullWidth size="small" sx={{ mb: 2 }}>
                    <InputLabel>Тип действия</InputLabel>
                    <Select
                        value={actionType}
                        onChange={(e) => handleActionTypeChange(e.target.value as ActionType)}
                        label="Тип действия"
                    >
                        {Object.entries(actionLabels).map(([value, label]) => (
                            <MenuItem key={value} value={value}>
                                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                    {actionIcons[value as ActionType]}
                                    {label}
                                </Box>
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>

                <Box sx={{ minHeight: 150 }}>
                    {renderActionConfig()}
                </Box>
            </Box>

            <Handle
                type="source"
                position={Position.Bottom}
                style={{ background: '#555' }}
            />
        </Paper>
    );
};

export const ActionNode = memo(observer(ActionNodeComponent));