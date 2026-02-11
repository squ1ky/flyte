import { api } from "../instance";
import type { Passenger, AddPassengerRequest, AddPassengerResponse } from "../../types/user";

export const userApi = {
    getPassengers: async (userId: number) => {
        const response = await api.get<Passenger[]>(`/users/${userId}/passengers`);
        return response.data;
    },

    addPassenger: async (userId: number, data: AddPassengerRequest) => {
        const response = await api.post<AddPassengerResponse>(`/users/${userId}/passengers`, data);
        return response.data;
    },
};
