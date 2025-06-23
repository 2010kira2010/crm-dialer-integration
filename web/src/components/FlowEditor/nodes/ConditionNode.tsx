import React, { memo, useState } from 'react';
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
} from '@mui/material';
import { useStores } from '../../../hooks/useStores';

export const ConditionNode = memo(({ data, id }: any) => {
    const { dataStore } = useStores();
    const [field, setField] = useState(data.field || '');
    const [operator, setOperator] = useState(data.operator || 'equals');
    const [value, setValue] = useState(data.value || '');

    const operators = [
        { value: 'equals', label: 'Равно' },
        { value: 'not_equals', label: 'Не равно' },
        { value: 'greater_than', label: 'Больше' },
        { value: 'less_than', label: 'Меньше' },
        { value: 'contains', label: 'Содержит' },
    ];

    return (
        <Paper sx={{ p: 2, minWidth: 250 }}>
            <Handle type="target" position={Position.Top} />

            <Typography variant="subtitle2" gutterBottom>
                Условие
            </Typography>

            <Box sx={{ mt: 1 }}>
                <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                    <InputLabel>Поле</InputLabel>
                    <Select
                        value={field}
                        onChange={(e) => setField(e.target.value)}
                        label="Поле"
                    >
                        {dataStore.amocrmFields.map((f) => (
                            <MenuItem key={f.id} value={f.name}>
                                {f.name}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>

                <FormControl fullWidth size="small" sx={{ mb: 1 }}>
                    <InputLabel>Оператор</InputLabel>
                    <Select
                        value={operator}
                        onChange={(e) => setOperator(e.target.value)}
                        label="Оператор"
                    >
                        {operators.map((op) => (
                            <MenuItem key={op.value} value={op.value}>
                                {op.label}
                            </MenuItem>
                        ))}
                    </Select>
                </FormControl>

                <TextField
                    fullWidth
                    size="small"
                    label="Значение"
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                />
            </Box>

            <Handle
                type="source"
                position={Position.Bottom}
                id="true"
                style={{ left: '25%' }}
            />
            <Handle
                type="source"
                position={Position.Bottom}
                id="false"
                style={{ left: '75%' }}
            />
        </Paper>
    );
});