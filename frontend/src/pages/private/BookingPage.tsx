import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useQuery, useMutation } from "@tanstack/react-query";
import { ArrowLeft, CreditCard, Loader } from "lucide-react";
import { Header } from "../../components/layout/Header";
import { flightApi } from "../../api/endpoints/flight";
import { userApi } from "../../api/endpoints/user";
import { bookingApi } from "../../api/endpoints/booking";
import { useAuthStore } from "../../store/authStore";
import { PassengerSelector } from "../../features/booking/PassengerSelector";
import { SeatMap } from "../../features/booking/SeatMap";
import { formatPrice } from "../../lib/format";

export const BookingPage = () => {
    const { flightId } = useParams();
    const navigate = useNavigate();
    const { userId } = useAuthStore();

    const [selectedPassengerIds, setSelectedPassengerIds] = useState<number[]>([]);
    const [selectedSeats, setSelectedSeats] = useState<Record<number, string>>({});
    const [currentPassengerId, setCurrentPassengerId] = useState<number | null>(null);

    const { data: flight, isLoading: flightLoading } = useQuery({
        queryKey: ["flight", flightId],
        queryFn: () => flightApi.getFlight(flightId!),
        enabled: !!flightId,
    });

    const { data: passengers } = useQuery({
        queryKey: ["passengers", userId],
        queryFn: () => userApi.getPassengers(userId!),
        enabled: !!userId,
    });

    const { data: seats } = useQuery({
        queryKey: ["seats", flightId],
        queryFn: () => bookingApi.getAvailableSeats(Number(flightId)),
        enabled: !!flightId,
    });

    const bookingMutation = useMutation({
        mutationFn: bookingApi.createBooking,
    });

    const handleTogglePassenger = (id: number) => {
        if (selectedPassengerIds.includes(id)) {
            setSelectedPassengerIds((prev) => prev.filter((p) => p !== id));
            if (currentPassengerId === id) setCurrentPassengerId(null);

            const newSeats = { ...selectedSeats };
            delete newSeats[id];
            setSelectedSeats(newSeats);
        } else {
            setSelectedPassengerIds((prev) => [...prev, id]);
            setCurrentPassengerId(id);
        }
    };

    const handleSelectSeat = (seatNumber: string) => {
        if (!currentPassengerId) return;

        const oldOwnerId = Object.keys(selectedSeats).find(key => selectedSeats[Number(key)] === seatNumber);
        const newSeats = { ...selectedSeats };
        if (oldOwnerId) delete newSeats[Number(oldOwnerId)];

        newSeats[currentPassengerId] = seatNumber;
        setSelectedSeats(newSeats);

        const nextPassenger = selectedPassengerIds.find(id => id !== currentPassengerId && !newSeats[id]);
        if (nextPassenger) {
            setCurrentPassengerId(nextPassenger);
        }
    };

    const handleDeselectSeat = (seatNumber: string) => {
        const passengerId = Number(Object.keys(selectedSeats).find(key => selectedSeats[Number(key)] === seatNumber));

        if (passengerId) {
            const newSeats = { ...selectedSeats };
            delete newSeats[passengerId];
            setSelectedSeats(newSeats);
            setCurrentPassengerId(passengerId);
        }
    };

    const handleBookAll = async () => {
        if (selectedPassengerIds.length === 0) {
            alert("Выберите хотя бы одного пассажира");
            return;
        }

        for (const pId of selectedPassengerIds) {
            if (!selectedSeats[pId]) {
                alert("Выберите место для каждого пассажира");
                return;
            }
        }

        try {
            for (const pId of selectedPassengerIds) {
                const passenger = passengers?.find(p => p.id === pId);
                if (!passenger) continue;

                const seat = seats?.find(s => s.seat_number === selectedSeats[pId]);
                const price = flight ? Math.round(flight.base_price_cents * (seat?.price_multiplier || 1)) : 0;

                await bookingMutation.mutateAsync({
                    flight_id: Number(flightId),
                    seat_number: selectedSeats[pId],
                    passenger_name: `${passenger.last_name} ${passenger.first_name}`,
                    passenger_passport: passenger.document_number,
                    price_cents: price,
                    currency: "RUB"
                });
            }
            alert("Бронирование успешно! Ожидайте подтверждения оплаты.");
            navigate("/profile");
        } catch (error) {
            console.error(error);
            alert("Ошибка при бронировании. Попробуйте снова.");
        }
    };

    if (flightLoading) return <div className="min-h-screen flex items-center justify-center">Загрузка...</div>;
    if (!flight) return <div>Рейс не найден</div>;

    const totalPrice = selectedPassengerIds.reduce((sum, pId) => {
        const seatNum = selectedSeats[pId];
        const seat = seats?.find(s => s.seat_number === seatNum);
        return sum + (flight.base_price_cents * (seat?.price_multiplier || 1));
    }, 0);

    return (
        <div className="min-h-screen bg-gray-50 font-sans pb-20">
            <Header />

            <div className="pt-24 max-w-7xl mx-auto px-6">
                <button onClick={() => navigate(-1)} className="flex items-center gap-2 text-gray-500 hover:text-gray-900 mb-4">
                    <ArrowLeft className="w-5 h-5" /> Назад
                </button>

                <h1 className="text-3xl font-bold mb-6">Бронирование билета</h1>

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    <div className="space-y-6">
                        <PassengerSelector
                            passengers={passengers || []}
                            selectedIds={selectedPassengerIds}
                            currentPassengerId={currentPassengerId}
                            selectedSeats={selectedSeats}
                            onToggle={handleTogglePassenger}
                            onSelectActive={setCurrentPassengerId}
                        />

                        {selectedPassengerIds.length > 0 && (
                            <div className="bg-white p-4 rounded-2xl shadow-sm border border-gray-100">
                                <h4 className="font-bold text-sm mb-2">Выбранные места:</h4>
                                {selectedPassengerIds.map((pId) => {
                                    const p = passengers?.find((x) => x.id === pId);
                                    return (
                                        <div key={pId} className="text-sm flex justify-between py-1 border-b border-gray-50 last:border-0">
                                            <span>{p?.first_name}</span>
                                            <span className="font-bold text-blue-600">{selectedSeats[pId] || "—"}</span>
                                        </div>
                                    );
                                })}
                            </div>
                        )}
                    </div>

                    <div className="lg:col-span-2">
                        <SeatMap
                            seats={seats || []}
                            selectedSeats={selectedSeats}
                            currentPassengerId={currentPassengerId}
                            onSelectSeat={handleSelectSeat}
                            onDeselectSeat={handleDeselectSeat}
                        />

                        <div className="mt-6 bg-white p-6 rounded-3xl shadow-lg flex flex-col md:flex-row items-center justify-between gap-6">
                            <div>
                                <div className="text-gray-500 mb-1">Итоговая стоимость</div>
                                <div className="text-3xl font-bold text-gray-900">{formatPrice(totalPrice)}</div>
                            </div>
                            <button
                                onClick={handleBookAll}
                                disabled={bookingMutation.isPending || selectedPassengerIds.length === 0}
                                className="w-full md:w-auto px-8 py-4 bg-green-600 text-white font-bold rounded-xl hover:bg-green-700 transition flex items-center justify-center gap-2 shadow-lg shadow-green-500/20 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {bookingMutation.isPending ? <Loader className="animate-spin" /> : <CreditCard />}
                                Перейти к оплате
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
