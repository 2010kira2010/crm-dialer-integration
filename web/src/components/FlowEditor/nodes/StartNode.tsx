import React, { memo } from 'react';
import { Handle, Position } from 'react-flow-renderer';
import { Box, Paper, Typography } from '@mui/material';
import PlayCircleOutlineIcon from '@mui/icons-material/PlayCircleOutline';

export const StartNode = memo(({ data }: any) => {
    return (
        <Paper
            sx={{
                p: 2,
                minWidth: 150,
                textAlign: 'center',
                bgcolor: 'success.light',
                border: '2px solid',
                borderColor: 'success.main',
            }}
        >
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
                <PlayCircleOutlineIcon color="success" />
                <Typography variant="subtitle1" fontWeight="bold">
                    Старт
                </Typography>
            </Box>

            <Handle
                type="source"
                position={Position.Bottom}
                style={{ background: '#4caf50' }}
            />
        </Paper>
    );
});