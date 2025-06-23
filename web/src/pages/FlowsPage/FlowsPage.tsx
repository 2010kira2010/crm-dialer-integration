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
            // Create new flow
            const flow = {
                name: newFlowName,
                flow_data: {
                    nodes: [
                        {
                            id: 'start_1',
                            type: 'start',
                            data: { label: 'Start' },
                            position: { x: 250, y: 50 },
                        },
                        {
                            id: 'end_1',
                            type: 'end',
                            data: { label: 'End' },
                            position: { x: 250, y: 300 },
                        },
                    ],
                    edges: [],
                },
                is_active: false,
            };

            // TODO: Call API to create flow
            setCreateDialogOpen(false);
            setNewFlowName('');
            // Navigate to editor
            navigate('/flows/new');
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
        // TODO: Toggle flow status
        console.log('Toggle flow:', flowId, isActive);
    };

    const handleDeleteFlow = async (flowId: string) => {
        if (window.confirm('Вы уверены, что хотите удалить этот поток?')) {
            // TODO: Delete flow
            console.log('Delete flow:', flowId);
            handleMenuClose();
        }
    };

    const handleDuplicateFlow = async (flowId: string) => {
        // TODO: Duplicate flow
        console.log('Duplicate flow:', flowId);
        handleMenuClose();
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