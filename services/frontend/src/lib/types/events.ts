export interface PaginationMetadata {
    total_items: number;
    total_pages: number;
    current_page: number;
    limit: number;
}

export interface RawPlateEvent {
    ID: number;
    camera_id: string;
    camera_name: string;
    plate: string;
    is_manual: boolean;
    image_key: string;
    timestamp: string;
    suggestions?: string;
    system_event_id?: number;
}

export interface RawWeightEvent {
    ID: number;
    scale_id: string;
    weight: number;
    raw_payload: string;
    timestamp: string;
    system_event_id?: number;
}

export interface GateEvent {
    ID: number;
    gate_id: number;
    permit_id?: number;
    timestamp: string;
    plate_events: RawPlateEvent[];
    weight_events: RawWeightEvent[];
}

export interface SystemEvent {
    ID: number;
    type: string;
    source_id: string;
    payload: string;
    timestamp: string;
}

export interface ApiResponse<T> {
    data: T[];
    metadata: PaginationMetadata;
}
