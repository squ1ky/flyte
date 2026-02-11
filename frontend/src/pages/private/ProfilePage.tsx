import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, User as UserIcon, LogOut, CreditCard } from "lucide-react";
import { Header } from "../../components/layout/Header";
import { useAuthStore } from "../../store/authStore";
import { userApi } from "../../api/endpoints/user";
import { PassengerForm } from "../../features/profile/PassengerForm";
import type { AddPassengerRequest, Passenger } from "../../types/user";

export const ProfilePage = () => {
    const [isAdding, setIsAdding] = useState(false);
    const { userId, logout } = useAuthStore();
    const queryClient = useQueryClient();

    const { data: passengers, isLoading } = useQuery({
        queryKey: ['passengers', userId],
        queryFn: () => userApi.getPassengers(userId!),
        enabled: !!userId,
    });

    const addPassengerMutation = useMutation({
        mutationFn: (data: AddPassengerRequest) => userApi.addPassenger(userId!, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['passengers'] });
            setIsAdding(false);
        },
        onError: (error) => {
            console.error("Failed to add passenger:", error);
            alert("Ошибка при добавлении пассажира. Проверьте данные.");
        }
    });

    return (
        <div className="min-h-screen bg-gray-50 font-sans pb-20">
            <Header />

            <div className="pt-10 max-w-5xl mx-auto px-6">
                <div className="flex justify-between items-center mb-8">
                    <h1 className="text-3xl font-bold text-gray-900">Личный кабинет</h1>
                    <button
                        onClick={logout}
                        className="flex items-center gap-2 text-red-600 hover:bg-red-50 px-4 py-2 rounded-xl transition font-medium"
                    >
                        <LogOut className="w-5 h-5" />
                        Выйти
                    </button>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    <div className="hidden lg:block space-y-2">
                        <div className="bg-white p-2 rounded-2xl shadow-sm border border-gray-100 sticky top-24">
                            <button className="w-full text-left px-4 py-3 bg-blue-50 text-blue-700 font-bold rounded-xl flex items-center gap-3">
                                <UserIcon className="w-5 h-5" /> Пассажиры
                            </button>
                            <button className="w-full text-left px-4 py-3 text-gray-600 hover:bg-gray-50 font-medium rounded-xl flex items-center gap-3 transition">
                                <CreditCard className="w-5 h-5" /> Мои бронирования
                            </button>
                        </div>
                    </div>

                    <div className="lg:col-span-2 space-y-6">

                        {!isAdding && (
                            <button
                                onClick={() => setIsAdding(true)}
                                className="w-full border-2 border-dashed border-gray-300 rounded-2xl p-6 flex flex-col items-center justify-center text-gray-500 hover:border-blue-500 hover:text-blue-600 hover:bg-blue-50/50 transition gap-2 group cursor-pointer"
                            >
                                <div className="bg-gray-100 p-3 rounded-full group-hover:bg-blue-100 transition">
                                    <Plus className="w-6 h-6" />
                                </div>
                                <span className="font-bold">Добавить пассажира</span>
                            </button>
                        )}

                        {isAdding && (
                            <PassengerForm
                                onSubmit={(data) => addPassengerMutation.mutate(data)}
                                isLoading={addPassengerMutation.isPending}
                                onCancel={() => setIsAdding(false)}
                            />
                        )}

                        {isLoading ? (
                            <div className="text-center py-10 text-gray-400">Загрузка списка пассажиров...</div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {passengers && passengers.length > 0 ? (
                                    passengers.map((p: Passenger) => (
                                        <div key={p.id} className="bg-white p-5 rounded-2xl border border-gray-100 shadow-sm flex flex-col sm:flex-row sm:justify-between sm:items-center hover:shadow-md transition gap-4">
                                            <div>
                                                <div className="font-bold text-lg text-gray-900">
                                                    {p.last_name} {p.first_name} {p.middle_name}
                                                </div>
                                                <div className="text-sm text-gray-500 mt-1 flex flex-wrap gap-3 items-center">
                                                    <span className="flex items-center gap-1">
                                                        {new Date(p.birth_date).toLocaleDateString()}
                                                    </span>
                                                    <span className="hidden sm:inline w-[1px] h-3 bg-gray-300"></span>
                                                    <span className="uppercase font-mono bg-gray-50 px-1.5 rounded text-xs border border-gray-200">
                                                        {p.document_number}
                                                    </span>
                                                </div>
                                            </div>

                                            <div className="flex items-center gap-2 self-start sm:self-center">
                                                <div className="bg-blue-50 px-3 py-1 rounded-lg text-xs font-bold text-blue-700 uppercase">
                                                    {p.document_type === 'passport' ? 'Паспорт РФ' : 'Загран'}
                                                </div>
                                                <div className="bg-gray-100 px-3 py-1 rounded-lg text-xs font-bold text-gray-600 uppercase">
                                                    {p.citizenship}
                                                </div>
                                            </div>
                                        </div>
                                    ))
                                ) : (
                                    !isAdding && (
                                        <div className="text-center py-10 bg-white rounded-2xl border border-gray-100">
                                            <p className="text-gray-500">У вас пока нет сохраненных пассажиров</p>
                                        </div>
                                    )
                                )}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};
