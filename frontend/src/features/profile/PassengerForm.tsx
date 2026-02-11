import { useForm } from "react-hook-form";
import { User, Calendar, FileText, Globe, Save } from "lucide-react";
import type { AddPassengerRequest } from "../../types/user";

interface PassengerFormProps {
    onSubmit: (data: AddPassengerRequest) => void;
    isLoading: boolean;
    onCancel: () => void;
}

export const PassengerForm = ({ onSubmit, isLoading, onCancel }: PassengerFormProps) => {
    const { register, handleSubmit } = useForm<AddPassengerRequest>();

    return (
        <form onSubmit={handleSubmit(onSubmit)} className="bg-gray-50 p-6 rounded-2xl border border-gray-200 animate-in fade-in slide-in-from-top-4">
            <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                <User className="w-5 h-5 text-blue-600" />
                Новый пассажир
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                    <input
                        {...register("last_name", { required: true })}
                        placeholder="Фамилия"
                        className="w-full h-11 px-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900 placeholder:text-gray-400"
                    />
                </div>
                <div>
                    <input
                        {...register("first_name", { required: true })}
                        placeholder="Имя"
                        className="w-full h-11 px-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900 placeholder:text-gray-400"
                    />
                </div>
                <div className="md:col-span-2">
                    <input
                        {...register("middle_name")}
                        placeholder="Отчество (если есть)"
                        className="w-full h-11 px-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900 placeholder:text-gray-400"
                    />
                </div>

                <div className="relative">
                    <Calendar className="absolute left-3 top-3 w-5 h-5 text-gray-400 pointer-events-none" />
                    <input
                        {...register("birth_date", { required: true })}
                        type="date"
                        className="w-full h-11 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900"
                    />
                </div>
                <select
                    {...register("gender", { required: true })}
                    className="w-full h-11 px-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900"
                >
                    <option value="" disabled selected>Пол</option>
                    <option value="male">Мужской</option>
                    <option value="female">Женский</option>
                </select>

                <div className="relative">
                    <FileText className="absolute left-3 top-3 w-5 h-5 text-gray-400 pointer-events-none" />
                    <input
                        {...register("document_number", { required: true })}
                        placeholder="Номер документа"
                        className="w-full h-11 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900 placeholder:text-gray-400"
                    />
                </div>
                <select
                    {...register("document_type", { required: true })}
                    className="w-full h-11 px-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900"
                >
                    <option value="passport">Паспорт РФ</option>
                    <option value="international_passport">Загранпаспорт</option>
                </select>

                <div className="relative md:col-span-2">
                    <Globe className="absolute left-3 top-3 w-5 h-5 text-gray-400 pointer-events-none" />
                    <input
                        {...register("citizenship", { required: true, maxLength: 3 })}
                        placeholder="Гражданство (RUS)"
                        className="w-full h-11 pl-10 pr-4 rounded-xl border border-gray-200 bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900 placeholder:text-gray-400 uppercase"
                        maxLength={3}
                    />
                </div>
            </div>

            <div className="flex justify-end gap-3 mt-6 pt-4 border-t border-gray-200">
                <button
                    type="button"
                    onClick={onCancel}
                    disabled={isLoading}
                    className="px-4 py-2 text-gray-600 hover:bg-gray-200 rounded-xl transition font-medium"
                >
                    Отмена
                </button>
                <button
                    type="submit"
                    disabled={isLoading}
                    className="px-6 py-2 bg-blue-600 text-white font-bold rounded-xl hover:bg-blue-700 transition flex items-center gap-2 shadow-lg shadow-blue-500/20 active:scale-95"
                >
                    <Save className="w-4 h-4" />
                    {isLoading ? "Сохранение..." : "Сохранить"}
                </button>
            </div>
        </form>
    );
};
