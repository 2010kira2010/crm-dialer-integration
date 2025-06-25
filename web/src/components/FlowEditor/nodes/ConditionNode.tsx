import React, { memo, useState, useEffect } from 'react';
import { Handle, Position } from 'react-flow-renderer';
import {
    Box,
    Paper,
    Typography,
    Select,
    MenuItem,
    TextField,
    FormControl,
    InputLabel,
    Chip,
} from '@mui/material';
import { useStores } from '../../../hooks/useStores';
import { observer } from 'mobx-react-lite';
import { ConditionData, ConditionOperator } from '../../../types';

interface ConditionNodeProps {
    data: any;
    id: string;
    selected?: boolean;
}

const ConditionNodeComponent: React.FC<ConditionNodeProps> = ({ data, id, selected }) => {
    const { dataStore, flowStore } = useStores();
    const [conditionData, setConditionData] = useState<ConditionData>(data.conditionData || {
        field: '',
        fieldType: 'amocrm_field',
        operator: 'equals',
        value: '',
    });

    useEffect(() => {
        if (dataStore.amocrmFields.length === 0) {
            dataStore.loadAllData();
        }
    }, [dataStore]);

    useEffect(() => {
        flowStore.updateNodeData(id, { conditionData });
    }, [conditionData, id, flowStore]);

    const operators: { value: ConditionOperator; label: string }[] = [
        { value: 'equals', label: 'Равно' },
        { value: 'not_equals', label: 'Не равно' },
        { value: 'greater_than', label: 'Больше' },
        { value: 'less_than', label: 'Меньше' },
        { value: 'contains', label: 'Содержит' },
    ];

    const fieldTypes = [
        { value: 'amocrm_field', label: 'Поле AmoCRM' },
        { value: 'pipeline', label: 'Воронка' },
        { value: 'status', label: 'Статус/Этап' },
        { value: 'bucket', label: 'Бакет' },
        { value: 'scheduler', label: 'Шедуллер' },
        { value: 'scheduler_step', label: 'Шаг шедуллера' },
        { value: 'dial_attempts', label: 'Попытки дозвона' },
    ];

    const updateCondition = (key: keyof ConditionData, value: any) => {
        setConditionData(prev => ({ ...prev, [key]: value }));
    };

    const getFieldOptions = () => {
        switch (conditionData.fieldType) {
            case 'amocrm_field':
                return dataStore.amocrmFields.map(f => (
                    <MenuItem key={f.id} value={f.name}>
                        {f.name}
                    </MenuItem>
                ));
            case 'pipeline':
                return dataStore.amocrmPipelines.map(p => (
                    <MenuItem key={p.id} value={p.id}>
                        {p.name}
                    </MenuItem>
                ));
            case 'status':
                return dataStore.amocrmPipelines.flatMap(p =>
                    p.statuses.map(s => (
                        <MenuItem key={s.id} value={s.id}>
                            {p.name} / {s.name}
                        </MenuItem>
                    ))
                );
            case 'bucket':
                return dataStore.dialerBuckets.map(b => (
                    <MenuItem key={b.id} value={b.id}>
                        {b.name}
                    </MenuItem>
                ));
            case 'scheduler':
                return dataStore.dialerSchedulers.map(s => (
                    <MenuItem key={s.id} value={s.id}>
                        {s.name}
                    </MenuItem>
                ));
            case 'scheduler_step':
            case 'dial_attempts':
                return null; // These will use text input
            default:
                return null;
        }
    };

    const renderValueInput = () => {
        if (conditionData.fieldType === 'scheduler_step' || conditionData.fieldType === 'dial_attempts') {
            return (
                <TextField
                    fullWidth
                    size="small"
                    label="Значение"
                    type="number"
                    value={conditionData.value}
                    onChange={(e) => updateCondition('value', e.target.value)}
                />
            );
        }

        const options = getFieldOptions();
        if (options) {
            return (
                <FormControl fullWidth size="small">
                    <InputLabel>Значение</InputLabel>
                    <Select
                        value={conditionData.value}
                        onChange={(e) => updateCondition('value', e.target.value)}
                        label="Значение"
                    >
                        {options}
                    </Select>
                </FormControl>
            );
        }

        return (
            <TextField
                fullWidth
                size="small"
                label="Значение"
                value={conditionData.value}
                onChange={(e) => updateCondition('value', e.target.value)}
            />
        );
    };

    return (
        <Paper
            sx={{
                p: 2,
                minWidth: 300,
                border: selected ? '2px solid #1976d2' : '1px solid #ccc',
                bgcolor: selected ? 'action.hover' : 'background.paper',
            }}
        >
            <Handle type="target" position={Position.Top} />

            <Box sx={{ mb: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
                    <Typography variant="subtitle2" fontWeight="bold">
                        Условие
                    </Typography>
                    {selected && (
                        <Chip label="Выбрано" size="small" color="primary" />
                    )}
                </Box>

                <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                    <InputLabel>Тип поля</InputLabel>
                    <Select
                        value={conditionData.fieldType}
                        onChange={(e) => {
                            updateCondition('fieldType', e.target.value);
                            updateCondition('field', '');
                            updateCondition('value', '');
                        }}
                        label="Тип поля"
                    >
                        {fieldTypes.map(type => (
                            <MenuItem key={type.value} value={type.value}>
                                {type.label}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>

                {(conditionData.fieldType === 'scheduler_step' || conditionData.fieldType === 'dial_attempts') ? (
                    <TextField
                        fullWidth
                        size="small"
                        sx={{ mb: 1 }}
                        label="Поле"
                        value={conditionData.fieldType === 'scheduler_step' ? 'Шаг шедуллера' : 'Попытки дозвона'}
                        disabled
                    />
                ) : (
                    <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                        <InputLabel>Поле</InputLabel>
                        <Select
                            value={conditionData.field}
                            onChange={(e) => updateCondition('field', e.target.value)}
                            label="Поле"
                        >
                            {getFieldOptions()}
                        </Select>
                    </FormControl>
                )}

                <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                    <InputLabel>Оператор</InputLabel>
                    <Select
                        value={conditionData.operator}
                        onChange={(e) => updateCondition('operator', e.target.value as ConditionOperator)}
                        label="Оператор"
                    >
                        {operators.map((op) => (
                            <MenuItem key={op.value} value={op.value}>
                                {op.label}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>

                {renderValueInput()}
            </Box>

            <Box sx={{ position: 'relative', height: 20 }}>
                <Handle
                    type="source"
                    position={Position.Bottom}
                    id="true"
                    style={{ left: '30%', background: '#4caf50' }}
                />
                <Typography
                    variant="caption"
                    sx={{ position: 'absolute', left: '20%', bottom: -20, fontSize: 10 }}
                >
                    Да
                </Typography>

                <Handle
                    type="source"
                    position={Position.Bottom}
                    id="false"
                    style={{ left: '70%', background: '#f44336' }}
                />
                <Typography
                    variant="caption"
                    sx={{ position: 'absolute', left: '65%', bottom: -20, fontSize: 10 }}
                >
                    Нет
                </Typography>
            </Box>
        </Paper>
    );
};

export const ConditionNode = memo(observer(ConditionNodeComponent));