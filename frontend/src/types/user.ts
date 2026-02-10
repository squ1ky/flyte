export interface User {
    id: number;
    email: string;
    phone_number?: string;
}

export interface AuthResponse {
    token: string;
}

export interface Passenger {
    id?: number;
    first_name: string;
    last_name: string;
    middle_name?: string;
    birth_date: string;
    gender: string;
    document_number: string;
    document_type: string;
    citizenship: string;
}

export interface AddPassengerRequest {
    first_name: string;
    last_name: string;
    middle_name?: string;
    birth_date: string;
    gender: string;
    document_number: string;
    document_type: string;
    citizenship: string;
}

export interface AddPassengerResponse {
    passenger_id: number;
}

export interface SignUpRequest {
    email: string;
    password: string;
    phone_number?: string;
}

export interface SignUpResponse {
    user_id: number;
}

export interface IDResponse {
    id: number;
}