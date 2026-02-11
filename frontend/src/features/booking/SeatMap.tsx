import type { Seat } from "../../types/booking";
import { cn } from "../../lib/utils";

interface SeatMapProps {
    seats: Seat[];
    selectedSeats: Record<number, string>;
    currentPassengerId: number | null;
    onSelectSeat: (seatNumber: string) => void;
    onDeselectSeat: (seatNumber: string) => void;
}

export const SeatMap = ({ seats, selectedSeats, currentPassengerId, onSelectSeat, onDeselectSeat }: SeatMapProps) => {
    const seatToPassengerId = Object.entries(selectedSeats).reduce((acc, [pId, seatNum]) => {
        acc[seatNum] = Number(pId);
        return acc;
    }, {} as Record<string, number>);

    const sortedSeats = [...seats].sort((a, b) => {
        const rowA = parseInt(a.seat_number);
        const rowB = parseInt(b.seat_number);
        if (rowA !== rowB) return rowA - rowB;
        return a.seat_number.localeCompare(b.seat_number);
    });

    const handleSeatClick = (seatNumber: string) => {
        const ownerId = seatToPassengerId[seatNumber];

        if (ownerId) {
            onDeselectSeat(seatNumber);
        } else {
            if (currentPassengerId) {
                onSelectSeat(seatNumber);
            }
        }
    };

    return (
        <div className="bg-white rounded-3xl p-6 shadow-sm border border-gray-100">
            <div className="flex justify-between items-center mb-6">
                <h3 className="text-lg font-bold text-gray-900">Карта мест</h3>
                {!currentPassengerId && Object.keys(selectedSeats).length < Object.keys(selectedSeats).length + 1 && (
                    <span className="text-xs text-gray-400">Выберите пассажира слева</span>
                )}
            </div>

            {/* Легенда */}
            <div className="flex gap-4 justify-center mb-6 text-xs text-gray-500">
                <div className="flex items-center gap-1.5"><div className="w-4 h-4 bg-gray-200 rounded"></div> Занято</div>
                <div className="flex items-center gap-1.5"><div className="w-4 h-4 bg-white border-2 border-green-500 rounded"></div> Свободно</div>
                <div className="flex items-center gap-1.5"><div className="w-4 h-4 bg-blue-600 rounded"></div> Выбрано</div>
            </div>

            <div className="relative bg-gray-50 rounded-2xl p-6 overflow-x-auto min-h-[300px]">
                {/* Нос самолета */}
                <div className="absolute left-1/2 top-0 -translate-x-1/2 w-20 h-8 bg-gray-200 rounded-b-3xl mb-4 opacity-50" />

                <div className="grid grid-cols-6 gap-3 min-w-[340px] max-w-md mx-auto pt-8">
                    {sortedSeats.map((seat) => {
                        const isOccupiedByOthers = seat.is_booked;
                        const myOwnerId = seatToPassengerId[seat.seat_number];
                        const isSelectedByCurrent = myOwnerId === currentPassengerId;

                        const isDisabled = isOccupiedByOthers || (!myOwnerId && !currentPassengerId);

                        return (
                            <button
                                key={seat.id}
                                disabled={isDisabled}
                                onClick={() => handleSeatClick(seat.seat_number)}
                                className={cn(
                                    "h-10 rounded-lg font-bold text-xs transition-all flex items-center justify-center relative",
                                    isOccupiedByOthers
                                        ? "bg-gray-200 text-gray-400 cursor-not-allowed"
                                        : myOwnerId
                                            ? isSelectedByCurrent
                                                ? "bg-blue-600 text-white shadow-lg shadow-blue-500/30 scale-110 z-10"
                                                : "bg-blue-400 text-white opacity-90"
                                            : "bg-white border-2 border-green-500 text-green-700 hover:bg-green-50 hover:scale-105 active:scale-95"
                                )}
                                title={isOccupiedByOthers ? "Занято" : myOwnerId ? "Нажмите для отмены" : `Место ${seat.seat_number}`}
                            >
                                {seat.seat_number}
                                {/* Если выбрано - покажем маленькую точку-индикатор отмены при ховере */}
                                {myOwnerId && (
                                    <span className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full border border-white opacity-0 hover:opacity-100 transition-opacity" />
                                )}
                            </button>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};
