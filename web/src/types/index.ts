export interface AmoCRMField {
    id: number;
    name: string;
    type: string;
    created_at: string;
    updated_at: string;
}

export interface DialerScheduler {
    id: string;
    name: string;
    created_at: string;
    updated_at: string;
}

export interface DialerCampaign {
    id: string;
    name: string;
    created_at: string;
    updated_at: string;
}

export interface DialerBucket {
    id: string;
    campaign_id: string;
    name: string;
    created_at: string;
    updated_at: string;
}

export interface IntegrationFlow {
    id: string;
    name: string;
    flow_data: any;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface FlowNode {
    id: string;
    type: 'start' | 'condition' | 'action' | 'end';
    data: any;
    position: { x: number; y: number };
}

export interface FlowEdge {
    id: string;
    source: string;
    target: string;
    type?: string;
    animated?: boolean;
}