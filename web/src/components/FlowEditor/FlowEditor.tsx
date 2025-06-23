import React, { useCallback, useEffect } from 'react';
import ReactFlow, {
    Node,
    Edge,
    addEdge,
    Background,
    Controls,
    MiniMap,
    useNodesState,
    useEdgesState,
    Connection,
    NodeTypes,
} from 'react-flow-renderer';
import { observer } from 'mobx-react-lite';
import { Box, Paper } from '@mui/material';
import { useStores } from '../../hooks/useStores';
import { StartNode } from './nodes/StartNode';
import { ConditionNode } from './nodes/ConditionNode';
import { ActionNode } from './nodes/ActionNode';
import { EndNode } from './nodes/EndNode';
import { FlowToolbar } from './FlowToolbar';

const nodeTypes: NodeTypes = {
    start: StartNode,
    condition: ConditionNode,
    action: ActionNode,
    end: EndNode,
};

export const FlowEditor: React.FC = observer(() => {
    const { flowStore } = useStores();
    const [nodes, setNodes, onNodesChange] = useNodesState(flowStore.nodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(flowStore.edges);

    useEffect(() => {
        setNodes(flowStore.nodes);
        setEdges(flowStore.edges);
    }, [flowStore.nodes, flowStore.edges, setNodes, setEdges]);

    const onConnect = useCallback(
        (params: Edge | Connection) => {
            setEdges((eds) => addEdge(params, eds));
        },
        [setEdges]
    );

    const onNodeDelete = useCallback(
        (deleted: Node[]) => {
            deleted.forEach((node) => {
                flowStore.removeNode(node.id);
            });
        },
        [flowStore]
    );

    const onDragOver = useCallback((event: React.DragEvent) => {
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
    }, []);

    const onDrop = useCallback(
        (event: React.DragEvent) => {
            event.preventDefault();

            const type = event.dataTransfer.getData('application/reactflow');
            if (!type) return;

            const reactFlowBounds = event.currentTarget.getBoundingClientRect();
            const position = {
                x: event.clientX - reactFlowBounds.left,
                y: event.clientY - reactFlowBounds.top,
            };

            const newNode: Node = {
                id: `${type}_${Date.now()}`,
                type,
                position,
                data: { label: `${type} node` },
            };

            flowStore.addNode(newNode as any);
        },
        [flowStore]
    );

    return (
        <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
            <FlowToolbar />
            <Paper sx={{ flex: 1, m: 2 }}>
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    onNodesChange={onNodesChange}
                    onEdgesChange={onEdgesChange}
                    onConnect={onConnect}
                    onNodesDelete={onNodeDelete}
                    onDrop={onDrop}
                    onDragOver={onDragOver}
                    nodeTypes={nodeTypes}
                    fitView
                >
                    <Background />
                    <Controls />
                    <MiniMap />
                </ReactFlow>
            </Paper>
        </Box>
    );
});