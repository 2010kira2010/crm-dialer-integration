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

    clear() {
        this.flows = [];
        this.currentFlow = null;
        this.nodes = [];
        this.edges = [];
        this.isLoading = false;
        this.error = null;
    }
}