import UpdateIcon from "@mui/icons-material/Update";
import CallIcon from "@mui/icons-material/Call";
import PriorityHighIcon from "@mui/icons-material/PriorityHigh";
import ScheduleIcon from "@mui/icons-material/Schedule";
import RemoveCircleIcon from "@mui/icons-material/RemoveCircle";
import React from "react";

export type ConditionOperator =
    | 'equals'
    | 'not_equals'
    | 'greater_than'
    | 'less_than'
    | 'contains';

export type ActionType =
    | 'update_lead'
    | 'add_to_bucket'
    | 'change_priority'
    | 'change_scheduler_step'
    | 'remove_from_dialer';

export interface ConditionData {
    field: string;
    fieldType: string;
    value: string;
    operator: string;
}

export interface AmoCRMField {
    id: number;
    name: string;
    type: string;
    created_at: string;
    updated_at: string;
}

export interface AmoCRMPipeline {
    id: number;
    name: string;
    statuses: AmoCRMStatus[];
}

export interface AmoCRMStatus {
    id: number;
    name: string;
    pipeline_id: number;
    sort: number;
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