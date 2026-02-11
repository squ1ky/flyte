import { api } from "../instance";
import { useAuthStore } from "../../store/authStore";
import type { CreateBookingRequest, Booking, Seat } from "../../types/booking";

const getAuthHeaders = () => {
    const token = useAuthStore.getState().token;
    return token ? { Authorization: `Bearer ${token}` } : {};
};

export const bookingApi = {
    getAvailableSeats: async (flightId: number) => {
        const response = await api.get<Seat[]>(`/flights/${flightId}/seats`, {
            headers: getAuthHeaders()
        });
        return response.data;
    },

    createBooking: async (data: CreateBookingRequest) => {
        const response = await api.post<Booking>("/bookings", data, {
            headers: getAuthHeaders()
        });
        return response.data;
    },
};
