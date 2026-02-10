import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { Header } from "../../components/layout/Header";
import { SearchForm } from "../../features/flight-search/SearchForm";
import { FlightCard } from "../../features/flight-search/FlightCard";
import { flightApi } from "../../api/endpoints/flight";
import type { SearchFlightParams } from "../../types/flight";

export const HomePage = () => {
    const [searchParams, setSearchParams] = useState<SearchFlightParams | null>(null);

    const { data: flights, isLoading, isError } = useQuery({
        queryKey: ['flights', searchParams],
        queryFn: () => searchParams ? flightApi.searchFlights(searchParams) : Promise.resolve([]),
        enabled: !!searchParams,
        staleTime: 0,
    });

    const handleSearch = (params: SearchFlightParams) => {
        setSearchParams(params);
    };

    return (
        <div className="min-h-screen bg-gray-50 font-sans text-gray-900">
            <Header />

            <div className="relative pt-32 pb-20 px-6 md:pt-48 md:pb-32 bg-blue-600 text-white overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-br from-blue-600 via-blue-700 to-indigo-800" />

                <div className="relative z-10 max-w-6xl mx-auto flex flex-col items-center text-center">
                    <h1 className="text-4xl md:text-6xl font-extrabold tracking-tight mb-6 drop-shadow-lg">
                        Поиск дешевых авиабилетов
                    </h1>

                    <SearchForm onSearch={handleSearch} isLoading={isLoading} />
                </div>
            </div>

            <div className="max-w-4xl mx-auto px-6 -mt-10 relative z-20 mb-20">
                {isError && (
                    <div className="bg-red-50 text-red-600 p-6 rounded-2xl shadow-lg border border-red-100 text-center">
                        Ошибка поиска. Повторите попытку.
                    </div>
                )}

                {isLoading && (
                    <div className="space-y-4">
                        {[1, 2, 3].map(i => (
                            <div key={i} className="h-40 bg-white rounded-2xl shadow-sm animate-pulse" />
                        ))}
                    </div>
                )}

                {!isLoading && flights && (
                    <div className="space-y-4">
                        {flights.length > 0 ? (
                            flights.map((flight) => (
                                <FlightCard key={flight.id} flight={flight} />
                            ))
                        ) : (
                            <div className="bg-white p-10 rounded-2xl shadow-lg text-center">
                                <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
                                    <Search className="h-8 w-8 text-gray-400" />
                                </div>
                                <h3 className="text-xl font-bold text-gray-900">Рейсов не найдено</h3>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};
