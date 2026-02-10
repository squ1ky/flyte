import { api } from "../instance";
import type {AuthResponse, SignUpRequest, SignUpResponse} from "../../types/user";

interface LoginRequest {
    email: string;
    password: string;
}

export const authApi = {
    register: async (data: SignUpRequest) => {
        const response = await api.post<SignUpResponse>("/auth/sign-up", data);
        return response.data;
    },

    login: async (data: LoginRequest) => {
        const response = await api.post<AuthResponse>("/auth/sign-in", data);
        return response.data;
    },
};
