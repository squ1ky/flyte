import { api } from "../instance";
import type {Airport, Flight, SearchFlightParams} from "../../types/flight";

export const flightApi = {
    getAirports: async () => {
        const response = await api.get<Airport[]>("/airports");
        return response.data;
    },

    searchFlights: async (params: SearchFlightParams) => {
        const response = await api.get<{ flights: Flight[] }>("/flights", { params });
        return response.data.flights || [];
    },

    getFlight: async (id: string) => {
        const response = await api.get<Flight>(`/flights/${id}`);
        return response.data;
    },

    getSeats: async (id: string) => {
        const response = await api.get<any>(`/flights/${id}/seats`);
        return response.data;
    }
};
