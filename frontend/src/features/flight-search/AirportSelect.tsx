import { useQuery } from "@tanstack/react-query";
import { MapPin } from "lucide-react";
import { flightApi } from "../../api/endpoints/flight";
import { cn } from "../../lib/utils";

interface AirportSelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
    label: string;
    className?: string;
}

export const AirportSelect = ({ label, className, ...props }: AirportSelectProps) => {
    const { data: airports, isLoading } = useQuery({
        queryKey: ['airports'],
        queryFn: flightApi.getAirports,
        staleTime: 1000 * 60 * 60,
    });

    return (
        <div className={cn("relative group flex-1", className)}>
            <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-blue-500 transition pointer-events-none">
                <MapPin className="h-5 w-5" />
            </div>

            <select
                {...props}
                disabled={isLoading}
                className="w-full h-14 pl-12 pr-10 rounded-xl bg-gray-100/50 hover:bg-gray-100 focus:bg-white focus:ring-2 focus:ring-blue-500 outline-none transition font-medium text-lg appearance-none cursor-pointer text-gray-900"
            >
                <option value="" disabled hidden>{isLoading ? "Загрузка..." : label}</option>
                {airports?.map((airport) => (
                    <option key={airport.id} value={airport.code} className="text-gray-900">
                        {airport.city} ({airport.code}) - {airport.name}
                    </option>
                ))}
            </select>

            <div className="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none">
                <svg className="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7"></path></svg>
            </div>
        </div>
    );
};
