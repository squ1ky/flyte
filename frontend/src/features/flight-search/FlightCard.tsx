import { Plane, MapPin } from "lucide-react";
import { formatPrice, formatTime, getFlightDuration } from "../../lib/format";
import type { Flight } from "../../types/flight";

interface FlightCardProps {
    flight: Flight;
    onSelect?: (id: number) => void;
}

export const FlightCard = ({ flight, onSelect }: FlightCardProps) => {
    return (
        <div className="bg-white rounded-3xl p-6 shadow-sm border border-gray-100 hover:shadow-lg transition-shadow duration-300 flex flex-col md:flex-row gap-6 relative group overflow-hidden">
            <div className="flex-1 flex flex-col justify-center">
                <div className="flex items-center justify-between gap-4">
                    <div className="text-left min-w-[80px]">
                        <div className="text-3xl font-bold text-gray-900 leading-none mb-1">
                            {formatTime(flight.departure_time)}
                        </div>
                        <div className="text-sm font-medium text-gray-500">
                            {flight.departure_airport}
                        </div>
                    </div>

                    <div className="flex-1 px-4 flex flex-col items-center relative min-w-[120px]">
                        <div className="text-xs text-gray-400 font-medium mb-1">
                            В пути: {getFlightDuration(flight.departure_time, flight.arrival_time)}
                        </div>

                        <div className="w-full relative h-6 flex items-center justify-center">
                            <div className="absolute inset-x-0 h-[2px] bg-gray-200" />
                            <div className="absolute left-0 w-2 h-2 rounded-full bg-gray-300 ring-2 ring-white" />
                            <div className="absolute right-0 w-2 h-2 rounded-full bg-gray-300 ring-2 ring-white" />

                            <div className="bg-white p-1 relative z-10 rotate-90 text-blue-500">
                                <Plane size={20} strokeWidth={2} fill="currentColor" className="text-blue-100" />
                                <Plane size={20} strokeWidth={2} className="absolute inset-1 text-blue-600" />
                            </div>
                        </div>

                        <div className="text-[10px] text-gray-400 mt-1 font-mono bg-gray-50 px-2 py-0.5 rounded-full border border-gray-100">
                            {flight.flight_number}
                        </div>
                    </div>

                    <div className="text-right min-w-[80px]">
                        <div className="text-3xl font-bold text-gray-900 leading-none mb-1">
                            {formatTime(flight.arrival_time)}
                        </div>
                        <div className="text-sm font-medium text-gray-500">
                            {flight.arrival_airport}
                        </div>
                    </div>
                </div>

                <div className="flex justify-between items-center mt-4 pt-4 border-t border-gray-50">
                    <div className="flex items-center gap-1.5 text-xs text-gray-400">
                        <MapPin className="w-3.5 h-3.5" />
                        <span>Терминал A</span>
                    </div>
                    {flight.available_seats !== undefined && flight.available_seats < 10 && (
                        <div className="text-xs text-red-500 font-medium bg-red-50 px-2 py-1 rounded-full">
                            Осталось {flight.available_seats} мест
                        </div>
                    )}
                </div>
            </div>

            <div className="w-full md:w-48 flex-shrink-0 flex flex-col justify-between border-t md:border-t-0 md:border-l border-gray-100 pt-6 md:pt-0 md:pl-6 gap-4">
                <div className="text-center md:text-right">
                    <div className="text-sm text-gray-400 mb-0.5">Стоимость</div>
                    <div className="text-2xl font-bold text-blue-600 tracking-tight">
                        {formatPrice(flight.base_price_cents)}
                    </div>
                </div>

                <button
                    onClick={() => onSelect?.(flight.id)}
                    className="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-500/20 active:scale-95 flex items-center justify-center gap-2"
                >
                    Выбрать
                </button>
            </div>
        </div>
    );
};
