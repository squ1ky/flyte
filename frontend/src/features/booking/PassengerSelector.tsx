import { User, Check, MousePointerClick } from "lucide-react";
import type { Passenger } from "../../types/user";
import { cn } from "../../lib/utils";

interface PassengerSelectorProps {
    passengers: Passenger[];
    selectedIds: number[];
    currentPassengerId: number | null;
    selectedSeats: Record<number, string>;
    onToggle: (id: number) => void;
    onSelectActive: (id: number) => void;
}

export const PassengerSelector = ({
                                      passengers,
                                      selectedIds,
                                      currentPassengerId,
                                      selectedSeats,
                                      onToggle,
                                      onSelectActive
                                  }: PassengerSelectorProps) => {
    return (
        <div className="bg-white rounded-3xl p-6 shadow-sm border border-gray-100">
            <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                <User className="w-5 h-5 text-blue-600" />
                Пассажиры
            </h3>

            <div className="space-y-3">
                {passengers.map((p) => {
                    const isSelected = selectedIds.includes(p.id!);
                    const isActive = currentPassengerId === p.id;
                    const seat = selectedSeats[p.id!];

                    return (
                        <div
                            key={p.id}
                            onClick={() => {
                                if (isSelected) onSelectActive(p.id!);
                                else onToggle(p.id!);
                            }}
                            className={cn(
                                "flex items-center gap-3 p-3 rounded-xl border-2 transition relative cursor-pointer group",
                                isActive
                                    ? "border-blue-500 bg-blue-50/50 shadow-md shadow-blue-100 ring-1 ring-blue-500"
                                    : isSelected
                                        ? "border-green-200 bg-white hover:border-blue-300"
                                        : "border-gray-100 opacity-60 hover:opacity-100 hover:border-gray-200"
                            )}
                        >
                            <div
                                onClick={(e) => { e.stopPropagation(); onToggle(p.id!); }}
                                className={cn(
                                    "w-6 h-6 rounded-lg flex items-center justify-center transition-colors shrink-0",
                                    isSelected ? "bg-green-500 text-white" : "bg-gray-100 text-gray-400 group-hover:bg-gray-200"
                                )}
                            >
                                {isSelected ? <Check className="w-4 h-4" /> : <User className="w-4 h-4" />}
                            </div>

                            <div className="flex-1 min-w-0">
                                <div className="font-bold text-gray-900 truncate">
                                    {p.last_name} {p.first_name}
                                </div>
                                <div className="text-xs text-gray-500 truncate">{p.document_number}</div>
                            </div>

                            {isSelected && (
                                <div className="flex items-center gap-2">
                                    {seat ? (
                                        <div className="bg-blue-600 text-white text-xs font-bold px-2 py-1 rounded-md shadow-sm">
                                            {seat}
                                        </div>
                                    ) : (
                                        isActive ? (
                                            <span className="text-xs text-blue-600 font-medium animate-pulse flex items-center gap-1">
                                                <MousePointerClick className="w-3 h-3" /> Выберите
                                            </span>
                                        ) : (
                                            <span className="text-xs text-orange-500 font-medium bg-orange-50 px-2 py-1 rounded-md">
                                                Нет места
                                            </span>
                                        )
                                    )}
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>

            {passengers.length === 0 && (
                <div className="text-center py-6 bg-gray-50 rounded-xl border border-dashed border-gray-200">
                    <p className="text-gray-500 text-sm">Нет пассажиров</p>
                </div>
            )}
        </div>
    );
};
