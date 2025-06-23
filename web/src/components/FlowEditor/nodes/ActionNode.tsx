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
} from '@mui/material';
import CallIcon from '@mui/icons-material/Call';
import UpdateIcon from '@mui/icons-material/Update';
import NoteAddIcon from '@mui/icons-material/NoteAdd';
import NotificationsIcon from '@mui/icons-material/Notifications';
import { useStores } from '../../../hooks/useStores';
import { observer } from 'mobx-react-lite';

const actionIcons: Record<string, React.ReactNode> = {
    send_to_dialer: <CallIcon />,
    update_lead: <UpdateIcon />,
    add_note: <NoteAddIcon />,
    send_notification: <NotificationsIcon />,
};

interface ActionNodeProps {
    data: any;
    id: string;
    selected?: boolean;
}

interface ActionData {
    scheduler_id?: string;
    campaign_id?: string;
    bucket_id?: string;
    fields?: Record<string, any>;
    text?: string;
    notification_type?: string;
    recipient?: string;
    subject?: string;
    message?: string;
}

const ActionNodeComponent: React.FC<ActionNodeProps> = ({ data, id, selected }) => {
    const { dataStore, flowStore } = useStores();
    const [actionType, setActionType] = useState<string>(data.type || 'send_to_dialer');
    const [actionData, setActionData] = useState<ActionData>(data.actionData || {});

    useEffect(() => {
        // Load data if not already loaded
        if (dataStore.dialerSchedulers.length === 0) {
            dataStore.loadAllData();
        }
    }, [dataStore]);

    useEffect(() => {
        // Update node data when actionType or actionData changes
        const node = flowStore.nodes.find(n => n.id === id);
        if (node) {
            node.data = {
                ...node.data,
                type: actionType,
                actionData: actionData,
            };
        }
    }, [actionType, actionData, id, flowStore.nodes]);

    const handleActionTypeChange = (type: string) => {
        setActionType(type);
        setActionData({});
    };

    const updateActionData = (key: string, value: any) => {
        setActionData((prev: ActionData) => ({ ...prev, [key]: value }));
    };

    const renderActionConfig = () => {
        switch (actionType) {
            case 'send_to_dialer':
                return (
                    <>
                        <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                            <InputLabel>Расписание</InputLabel>
                            <Select
                                value={actionData.scheduler_id || ''}
                                onChange={(e) => updateActionData('scheduler_id', e.target.value)}
                                label="Расписание"
                            >
                                <MenuItem value="">
                                    <em>Не выбрано</em>
                                </MenuItem>
                                {dataStore.dialerSchedulers.map((scheduler) => (
                                    <MenuItem key={scheduler.id} value={scheduler.id}>
                                        {scheduler.name}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                            <InputLabel>Кампания</InputLabel>
                            <Select
                                value={actionData.campaign_id || ''}
                                onChange={(e) => updateActionData('campaign_id', e.target.value)}
                                label="Кампания"
                            >
                                <MenuItem value="">
                                    <em>Не выбрано</em>
                                </MenuItem>
                                {dataStore.dialerCampaigns.map((campaign) => (
                                    <MenuItem key={campaign.id} value={campaign.id}>
                                        {campaign.name}
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>

                        <FormControl fullWidth size="small">
                            <InputLabel>Корзина</InputLabel>
                            <Select
                                value={actionData.bucket_id || ''}
                                onChange={(e) => updateActionData('bucket_id', e.target.value)}
                                label="Корзина"
                                disabled={!actionData.campaign_id}
                            >
                                <MenuItem value="">
                                    <em>Не выбрано</em>
                                </MenuItem>
                                {dataStore.dialerBuckets
                                    .filter(b => !actionData.campaign_id || b.campaign_id === actionData.campaign_id)
                                    .map((bucket) => (
                                        <MenuItem key={bucket.id} value={bucket.id}>
                                            {bucket.name}
                                        </MenuItem>
                                    ))}
                            </Select>
                        </FormControl>
                    </>
                );

            case 'update_lead':
                return (
                    <>
                        <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                            Выберите поля для обновления:
                        </Typography>
                        {dataStore.amocrmFields.slice(0, 5).map((field) => (
                            <Box key={field.id} sx={{ mb: 1 }}>
                                <TextField
                                    fullWidth
                                    size="small"
                                    label={field.name}
                                    value={actionData.fields?.[field.id] || ''}
                                    onChange={(e) => {
                                        const fields = { ...actionData.fields, [field.id]: e.target.value };
                                        updateActionData('fields', fields);
                                    }}
                                    helperText={`Тип: ${field.type}`}
                                />
                            </Box>
                        ))}
                    </>
                );

            case 'add_note':
                return (
                    <TextField
                        fullWidth
                        size="small"
                        label="Текст примечания"
                        multiline
                        rows={3}
                        value={actionData.text || ''}
                        onChange={(e) => updateActionData('text', e.target.value)}
                        helperText="Используйте {{field_name}} для вставки значений полей"
                    />
                );

            case 'send_notification':
                return (
                    <>
                        <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                            <InputLabel>Тип уведомления</InputLabel>
                            <Select
                                value={actionData.notification_type || 'email'}
                                onChange={(e) => updateActionData('notification_type', e.target.value)}
                                label="Тип уведомления"
                            >
                                <MenuItem value="email">Email</MenuItem>
                                <MenuItem value="sms">SMS</MenuItem>
                                <MenuItem value="push">Push</MenuItem>
                                <MenuItem value="telegram">Telegram</MenuItem>
                            </Select>
                        </FormControl>

                        <TextField
                            fullWidth
                            size="small"
                            label="Получатель"
                            value={actionData.recipient || ''}
                            onChange={(e) => updateActionData('recipient', e.target.value)}
                            sx={{ mb: 1 }}
                            helperText="Email, телефон или ID пользователя"
                        />

                        <TextField
                            fullWidth
                            size="small"
                            label="Тема (для email)"
                            value={actionData.subject || ''}
                            onChange={(e) => updateActionData('subject', e.target.value)}
                            sx={{ mb: 1 }}
                            disabled={actionData.notification_type !== 'email'}
                        />

                        <TextField
                            fullWidth
                            size="small"
                            label="Сообщение"
                            multiline
                            rows={3}
                            value={actionData.message || ''}
                            onChange={(e) => updateActionData('message', e.target.value)}
                            helperText="Используйте {{field_name}} для вставки значений"
                        />
                    </>
                );

            default:
                return null;
        }
    };

    return (
        <Paper
            sx={{
                p: 2,
                minWidth: 300,
                maxWidth: 400,
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
                        onChange={(e) => handleActionTypeChange(e.target.value)}
                        label="Тип действия"
                    >
                        <MenuItem value="send_to_dialer">
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <CallIcon fontSize="small" />
                                Отправить в автообзвон
                            </Box>
                        </MenuItem>
                        <MenuItem value="update_lead">
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <UpdateIcon fontSize="small" />
                                Обновить сделку
                            </Box>
                        </MenuItem>
                        <MenuItem value="add_note">
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <NoteAddIcon fontSize="small" />
                                Добавить примечание
                            </Box>
                        </MenuItem>
                        <MenuItem value="send_notification">
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <NotificationsIcon fontSize="small" />
                                Отправить уведомление
                            </Box>
                        </MenuItem>
                    </Select>
                </FormControl>

                <Box sx={{ minHeight: 200 }}>
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

// Правильный порядок оборачивания: сначала observer, потом memo
export const ActionNode = memo(observer(ActionNodeComponent));