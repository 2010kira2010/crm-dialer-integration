import React, { memo } from 'react';
import { Handle, Position } from 'react-flow-renderer';
import { Box, Paper, Typography } from '@mui/material';
import StopCircleIcon from '@mui/icons-material/StopCircle';

export const EndNode = memo(({ data }: any) => {
    return (
        <Paper
            sx={{
                p: 2,
                minWidth: 150,
                textAlign: 'center',
                bgcolor: 'error.light',
                border: '2px solid',
                borderColor: 'error.main',
            }}
        >
            <Handle
                type="target"
                position={Position.Top}
                style={{ background: '#f44336' }}
            />

            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
                <StopCircleIcon color="error" />
                <Typography variant="subtitle1" fontWeight="bold">
                    Конец
                </Typography>
            </Box>
        </Paper>
    );
});
