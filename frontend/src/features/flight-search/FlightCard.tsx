import { Plane, MapPin } from "lucide-react";
import { formatPrice, formatTime, getFlightDuration } from "../../lib/format";
import type { Flight } from "../../types/flight";

interface FlightCardProps {
    flight: Flight;
    onSelect?: (id: number) => void;
}

export const FlightCard = ({ flight, onSelect }: FlightCardProps) => {
    return (
        <div className="bg-white p-6 rounded-2xl shadow-lg border border-gray-100 hover:shadow-xl transition flex flex-col md:flex-row items-center justify-between gap-6 group">
            <div className="flex-1 w-full">
                <div className="flex items-center justify-between mb-4">
                    <div className="font-bold text-2xl text-gray-900">
                        {formatTime(flight.departure_time)}
                    </div>

                    <div className="flex-1 px-4 flex flex-col items-center relative">
                        <div className="text-xs text-gray-400 mb-1">
                            В пути: {getFlightDuration(flight.departure_time, flight.arrival_time)}
                        </div>
                        <div className="w-full h-[2px] bg-gray-200 relative">
                            <div className="absolute right-0 top-1/2 -translate-y-1/2 w-2 h-2 rounded-full bg-gray-300" />
                            <div className="absolute left-0 top-1/2 -translate-y-1/2 w-2 h-2 rounded-full bg-gray-300" />
                            <Plane className="absolute left-1/2 top-1/2 -translate-y-1/2 -translate-x-1/2 w-5 h-5 text-blue-500 rotate-90 bg-white p-[2px]" />
                        </div>
                        <div className="text-[10px] text-gray-400 mt-1 font-medium bg-gray-50 px-2 py-0.5 rounded-full">
                            {flight.flight_number}
                        </div>
                    </div>

                    <div className="font-bold text-2xl text-gray-900">
                        {formatTime(flight.arrival_time)}
                    </div>
                </div>

                <div className="flex justify-between text-gray-500 text-sm">
                    <div className="text-left">
                        <span className="font-bold text-lg text-gray-900 block leading-none mb-1">
                            {flight.departure_airport}
                        </span>
                        <div className="text-xs text-gray-500 flex items-center gap-1">
                            <MapPin className="w-3 h-3" /> Вылет
                        </div>
                    </div>
                    <div className="text-right">
                        <span className="font-bold text-lg text-gray-900 block leading-none mb-1">
                            {flight.arrival_airport}
                        </span>
                        <div className="text-xs text-gray-500 flex items-center justify-end gap-1">
                            Прилет <MapPin className="w-3 h-3" />
                        </div>
                    </div>
                </div>
            </div>

            <div className="w-full md:w-auto border-t md:border-t-0 md:border-l border-gray-100 pt-4 md:pt-0 md:pl-6 flex md:flex-col items-center justify-center gap-3 min-w-[140px]">
                <div className="text-2xl font-bold text-blue-600">
                    {formatPrice(flight.base_price_cents)}
                </div>

                {flight.available_seats !== undefined && flight.available_seats < 10 && (
                    <div className="text-xs text-red-500 font-medium">
                        Осталось {flight.available_seats} мест
                    </div>
                )}

                <button
                    onClick={() => onSelect?.(flight.id)}
                    className="px-6 py-3 rounded-xl bg-blue-600 text-white font-bold hover:bg-blue-700 transition w-full shadow-md shadow-blue-200"
                >
                    Выбрать
                </button>
            </div>
        </div>
    );
};
