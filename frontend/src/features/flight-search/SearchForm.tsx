import { useForm } from "react-hook-form";
import { Search, Calendar as CalendarIcon, Users } from "lucide-react";
import { AirportSelect } from "./AirportSelect";
import { cn } from "../../lib/utils";
import type { SearchFlightParams } from "../../types/flight";

interface SearchFormProps {
    onSearch: (params: SearchFlightParams) => void;
    isLoading?: boolean; // Добавим индикатор загрузки на кнопку
}

export const SearchForm = ({ onSearch, isLoading }: SearchFormProps) => {
    const { register, handleSubmit, formState: { errors } } = useForm<SearchFlightParams>({
        defaultValues: {
            passengers: 1,
            date: new Date(Date.now() + 86400000).toISOString().split('T')[0]
        }
    });

    const onSubmit = (data: SearchFlightParams) => {
        onSearch(data);
    };

    return (
        <form
            onSubmit={handleSubmit(onSubmit)}
            className="w-full max-w-5xl bg-white rounded-2xl p-2 shadow-2xl flex flex-col md:flex-row gap-2 relative z-20 text-left"
        >
            <AirportSelect
                label="Откуда"
                {...register("from", { required: true })}
                className={errors.from ? "ring-2 ring-red-500 rounded-xl" : ""}
            />

            <AirportSelect
                label="Куда"
                {...register("to", { required: true })}
                className={errors.to ? "ring-2 ring-red-500 rounded-xl" : ""}
            />

            <div className="flex-[0.6] relative group">
                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-blue-500 transition pointer-events-none">
                    <CalendarIcon className="h-5 w-5" />
                </div>
                <input
                    type="date"
                    {...register("date", { required: true })}
                    className={cn(
                        "w-full h-14 pl-12 pr-4 rounded-xl bg-gray-100/50 hover:bg-gray-100 focus:bg-white focus:ring-2 focus:ring-blue-500 outline-none transition font-medium text-lg text-gray-900",
                        errors.date && "ring-2 ring-red-500"
                    )}
                />
            </div>

            <div className="w-full md:w-32 relative group">
                <div className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-blue-500 transition pointer-events-none">
                    <Users className="h-5 w-5" />
                </div>
                <input
                    type="number"
                    min={1}
                    max={9}
                    {...register("passengers", { required: true, min: 1 })}
                    className="w-full h-14 pl-12 pr-4 rounded-xl bg-gray-100/50 hover:bg-gray-100 focus:bg-white focus:ring-2 focus:ring-blue-500 outline-none transition font-medium text-lg text-gray-900"
                />
            </div>

            <button
                type="submit"
                disabled={isLoading}
                className={cn(
                    "h-14 px-8 rounded-xl bg-orange-500 text-white font-bold text-lg",
                    "hover:bg-orange-600 active:scale-95 transition shadow-lg shadow-orange-500/30",
                    "flex items-center justify-center gap-2 md:w-auto w-full",
                    isLoading && "opacity-70 cursor-not-allowed"
                )}
            >
                {isLoading ? (
                    <span className="animate-spin">⌛</span> // Или любой спиннер
                ) : (
                    <Search className="h-5 w-5" />
                )}
                <span className="md:hidden">{isLoading ? "Поиск..." : "Найти билеты"}</span>
            </button>
        </form>
    );
};