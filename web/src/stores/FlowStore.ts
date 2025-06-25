import { makeAutoObservable, runInAction } from 'mobx';
import { FlowNode, FlowEdge, IntegrationFlow } from '../types';
import { RootStore } from './RootStore';
import api from '../services/api';

export class FlowStore {
    rootStore: RootStore;
    flows: IntegrationFlow[] = [];
    currentFlow: IntegrationFlow | null = null;
    nodes: FlowNode[] = [];
    edges: FlowEdge[] = [];
    isLoading = false;
    isSaving = false;
    error: string | null = null;

    constructor(rootStore: RootStore) {
        this.rootStore = rootStore;
        makeAutoObservable(this);
    }

    async loadFlows() {
        this.isLoading = true;
        this.error = null;
        try {
            const response = await api.get('/api/v1/flows');
            runInAction(() => {
                this.flows = response.data;
                this.isLoading = false;
            });
        } catch (error) {
            runInAction(() => {
                this.error = 'Failed to load flows';
                this.isLoading = false;
            });
        }
    }

    async loadFlow(id: string) {
        this.isLoading = true;
        this.error = null;
        try {
            const response = await api.get(`/api/v1/flows/${id}`);
            runInAction(() => {
                this.currentFlow = response.data;
                if (this.currentFlow?.flow_data) {
                    this.nodes = this.currentFlow.flow_data.nodes || [];
                    this.edges = this.currentFlow.flow_data.edges || [];
                }
                this.isLoading = false;
            });
        } catch (error) {
            runInAction(() => {
                this.error = 'Failed to load flow';
                this.isLoading = false;
            });
        }
    }

    async createFlow(name: string) {
        this.isLoading = true;
        this.error = null;
        try {
            const flowData = {
                name,
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
            };

            const response = await api.post('/api/v1/flows', flowData);
            runInAction(() => {
                this.currentFlow = response.data;
                this.isLoading = false;
            });
            return response.data.id;
        } catch (error) {
            runInAction(() => {
                this.error = 'Failed to create flow';
                this.isLoading = false;
            });
            throw error;
        }
    }

    async saveFlow() {
        if (!this.currentFlow) return;

        this.isLoading = true;
        this.error = null;
        try {
            const flowData = {
                ...this.currentFlow,
                flow_data: {
                    nodes: this.nodes,
                    edges: this.edges,
                },
            };

            const response = await api.put(`/api/v1/flows/${this.currentFlow.id}`, flowData);
            runInAction(() => {
                this.currentFlow = response.data;
                this.isLoading = false;
            });
        } catch (error) {
            runInAction(() => {
                this.error = 'Failed to save flow';
                this.isLoading = false;
            });
        }
    }

    async deleteFlow(id: string) {
        try {
            await api.delete(`/api/v1/flows/${id}`);
            runInAction(() => {
                this.flows = this.flows.filter(f => f.id !== id);
            });
        } catch (error) {
            runInAction(() => {
                this.error = 'Failed to delete flow';
            });
            throw error;
        }
    }

    async toggleFlow(id: string, isActive: boolean) {
        try {
            const flow = this.flows.find(f => f.id === id);
            if (!flow) return;

            const updatedFlow = { ...flow, is_active: isActive };
            const response = await api.put(`/api/v1/flows/${id}`, updatedFlow);

            runInAction(() => {
                const index = this.flows.findIndex(f => f.id === id);
                if (index !== -1) {
                    this.flows[index] = response.data;
                }
            });
        } catch (error) {
            runInAction(() => {
                this.error = 'Failed to toggle flow';
            });
            throw error;
        }
    }

    updateNodes(nodes: FlowNode[]) {
        this.nodes = nodes;
    }

    updateEdges(edges: FlowEdge[]) {
        this.edges = edges;
    }

    addNode(node: FlowNode) {
        this.nodes.push(node);
    }

    removeNode(nodeId: string) {
        this.nodes = this.nodes.filter(n => n.id !== nodeId);
        this.edges = this.edges.filter(e => e.source !== nodeId && e.target !== nodeId);
    }

    updateNodeData(nodeId: string, data: any) {
        const node = this.nodes.find(n => n.id === nodeId);
        if (node) {
            node.data = { ...node.data, ...data };
        }
    }

    clear() {
        this.flows = [];
        this.currentFlow = null;
        this.nodes = [];
        this.edges = [];
        this.isLoading = false;
        this.error = null;
    }
}