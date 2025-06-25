import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Box,
    Typography,
    Button,
    Grid,
    Card,
    CardContent,
    CardActions,
    IconButton,
    Chip,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
    Switch,
    FormControlLabel,
    Menu,
    MenuItem,
} from '@mui/material';
import {
    Add as AddIcon,
    Edit as EditIcon,
    Delete as DeleteIcon,
    MoreVert as MoreVertIcon,
    ContentCopy as ContentCopyIcon,
} from '@mui/icons-material';
import { observer } from 'mobx-react-lite';
import { useStores } from '../../hooks/useStores';

export const FlowsPage: React.FC = observer(() => {
    const navigate = useNavigate();
    const { flowStore } = useStores();
    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [newFlowName, setNewFlowName] = useState('');
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const [selectedFlow, setSelectedFlow] = useState<string | null>(null);

    useEffect(() => {
        flowStore.loadFlows();
    }, [flowStore]);

    const handleCreateFlow = async () => {
        if (newFlowName.trim()) {
            try {
                const flowId = await flowStore.createFlow(newFlowName);
                setCreateDialogOpen(false);
                setNewFlowName('');
                navigate(`/flows/${flowId}`);
            } catch (error) {
                console.error('Failed to create flow:', error);
            }
        }
    };

    const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, flowId: string) => {
        setAnchorEl(event.currentTarget);
        setSelectedFlow(flowId);
    };

    const handleMenuClose = () => {
        setAnchorEl(null);
        setSelectedFlow(null);
    };

    const handleToggleFlow = async (flowId: string, isActive: boolean) => {
        try {
            await flowStore.toggleFlow(flowId, isActive);
        } catch (error) {
            console.error('Failed to toggle flow:', error);
        }
    };

    const handleDeleteFlow = async (flowId: string) => {
        if (window.confirm('Вы уверены, что хотите удалить этот поток?')) {
            try {
                await flowStore.deleteFlow(flowId);
                handleMenuClose();
            } catch (error) {
                console.error('Failed to delete flow:', error);
            }
        }
    };

    const handleDuplicateFlow = async (flowId: string) => {
        const flow = flowStore.flows.find(f => f.id === flowId);
        if (flow) {
            try {
                const newFlowId = await flowStore.createFlow(`${flow.name} (копия)`);
                // TODO: Copy flow data
                handleMenuClose();
                navigate(`/flows/${newFlowId}`);
            } catch (error) {
                console.error('Failed to duplicate flow:', error);
            }
        }
    };

    return (
        <Box>
            <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="h4">Потоки интеграции</Typography>
                <Button
                    variant="contained"
                    startIcon={<AddIcon />}
                    onClick={() => setCreateDialogOpen(true)}
                >
                    Создать поток
                </Button>
            </Box>

            <Grid container spacing={3}>
                {flowStore.flows.map((flow) => (
                    <Grid item xs={12} sm={6} md={4} key={flow.id}>
                        <Card>
                            <CardContent>
                                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                                    <Box>
                                        <Typography variant="h6" gutterBottom>
                                            {flow.name}
                                        </Typography>
                                        <Chip
                                            label={flow.is_active ? 'Активен' : 'Неактивен'}
                                            color={flow.is_active ? 'success' : 'default'}
                                            size="small"
                                        />
                                    </Box>
                                    <IconButton
                                        size="small"
                                        onClick={(e) => handleMenuOpen(e, flow.id)}
                                    >
                                        <MoreVertIcon />
                                    </IconButton>
                                </Box>
                                <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                                    Создан: {new Date(flow.created_at).toLocaleDateString()}
                                </Typography>
                                <Typography variant="body2" color="text.secondary">
                                    Обновлен: {new Date(flow.updated_at).toLocaleDateString()}
                                </Typography>
                            </CardContent>
                            <CardActions>
                                <Button
                                    size="small"
                                    startIcon={<EditIcon />}
                                    onClick={() => navigate(`/flows/${flow.id}`)}
                                >
                                    Редактировать
                                </Button>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={flow.is_active}
                                            onChange={(e) => handleToggleFlow(flow.id, e.target.checked)}
                                            size="small"
                                        />
                                    }
                                    label=""
                                />
                            </CardActions>
                        </Card>
                    </Grid>
                ))}
            </Grid>

            {/* Create Flow Dialog */}
            <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)}>
                <DialogTitle>Создать новый поток</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="Название потока"
                        fullWidth
                        variant="outlined"
                        value={newFlowName}
                        onChange={(e) => setNewFlowName(e.target.value)}
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setCreateDialogOpen(false)}>Отмена</Button>
                    <Button onClick={handleCreateFlow} variant="contained">
                        Создать
                    </Button>
                </DialogActions>
            </Dialog>

            {/* Flow Actions Menu */}
            <Menu
                anchorEl={anchorEl}
                open={Boolean(anchorEl)}
                onClose={handleMenuClose}
            >
                <MenuItem onClick={() => handleDuplicateFlow(selectedFlow!)}>
                    <ContentCopyIcon fontSize="small" sx={{ mr: 1 }} />
                    Дублировать
                </MenuItem>
                <MenuItem onClick={() => handleDeleteFlow(selectedFlow!)}>
                    <DeleteIcon fontSize="small" sx={{ mr: 1 }} />
                    Удалить
                </MenuItem>
            </Menu>
        </Box>
    );
});