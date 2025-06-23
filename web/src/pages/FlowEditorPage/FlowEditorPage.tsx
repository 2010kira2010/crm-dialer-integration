import React, { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Box, CircularProgress, Alert } from '@mui/material';
import { observer } from 'mobx-react-lite';
import { useStores } from '../../hooks/useStores';
import { FlowEditor } from '../../components/FlowEditor/FlowEditor';

export const FlowEditorPage: React.FC = observer(() => {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { flowStore } = useStores();

    useEffect(() => {
        if (id && id !== 'new') {
            flowStore.loadFlow(id);
        } else if (id === 'new') {
            // Initialize new flow
            flowStore.currentFlow = {
                id: '',
                name: 'Новый поток',
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
                            position: { x: 250, y: 400 },
                        },
                    ],
                    edges: [],
                },
                is_active: false,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
            };
            flowStore.nodes = flowStore.currentFlow.flow_data.nodes;
            flowStore.edges = flowStore.currentFlow.flow_data.edges;
        }
    }, [id, flowStore]);

    if (flowStore.isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
                <CircularProgress />
            </Box>
        );
    }

    if (flowStore.error) {
        return (
            <Box sx={{ p: 3 }}>
                <Alert severity="error" onClose={() => navigate('/flows')}>
                    {flowStore.error}
                </Alert>
            </Box>
        );
    }

    return <FlowEditor />;
});