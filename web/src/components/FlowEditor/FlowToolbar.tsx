import React from 'react';
import { observer } from 'mobx-react-lite';
import {
    Box,
    AppBar,
    Toolbar,
    Typography,
    Button,
    IconButton,
    Divider,
} from '@mui/material';
import SaveIcon from '@mui/icons-material/Save';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import StopIcon from '@mui/icons-material/Stop';
import { useStores } from '../../hooks/useStores';

export const FlowToolbar: React.FC = observer(() => {
    const { flowStore } = useStores();

    const handleSave = async () => {
        await flowStore.saveFlow();
    };

    const onDragStart = (event: React.DragEvent, nodeType: string) => {
        event.dataTransfer.setData('application/reactflow', nodeType);
        event.dataTransfer.effectAllowed = 'move';
    };

    return (
        <AppBar position="static" color="default" elevation={1}>
            <Toolbar>
                <Typography variant="h6" sx={{ flexGrow: 0, mr: 3 }}>
                    {flowStore.currentFlow?.name || 'Новый поток'}
                </Typography>

                <Box sx={{ display: 'flex', gap: 2, flexGrow: 1 }}>
                    <Box
                        draggable
                        onDragStart={(e) => onDragStart(e, 'condition')}
                        sx={{
                            p: 1,
                            border: '1px dashed #ccc',
                            borderRadius: 1,
                            cursor: 'grab',
                            '&:hover': { bgcolor: 'action.hover' },
                        }}
                    >
                        <Typography variant="body2">Условие</Typography>
                    </Box>

                    <Box
                        draggable
                        onDragStart={(e) => onDragStart(e, 'action')}
                        sx={{
                            p: 1,
                            border: '1px dashed #ccc',
                            borderRadius: 1,
                            cursor: 'grab',
                            '&:hover': { bgcolor: 'action.hover' },
                        }}
                    >
                        <Typography variant="body2">Действие</Typography>
                    </Box>
                </Box>

                <Divider orientation="vertical" flexItem sx={{ mx: 2 }} />

                <Button
                    variant="contained"
                    startIcon={<SaveIcon />}
                    onClick={handleSave}
                    disabled={flowStore.isLoading}
                >
                    Сохранить
                </Button>

                {flowStore.currentFlow?.is_active ? (
                    <IconButton color="error" title="Остановить поток">
                        <StopIcon />
                    </IconButton>
                ) : (
                    <IconButton color="success" title="Запустить поток">
                        <PlayArrowIcon />
                    </IconButton>
                )}
            </Toolbar>
        </AppBar>
    );
});