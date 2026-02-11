import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { Plane, Mail, Lock, ArrowRight } from "lucide-react";
import { authApi } from "../../api/endpoints/auth";
import { useAuthStore } from "../../store/authStore";
import { cn } from "../../lib/utils";

interface LoginFormInputs {
    email: string;
    password: string;
}

export const LoginPage = () => {
    const navigate = useNavigate();
    const login = useAuthStore((state: { login: any; }) => state.login);
    const [serverError, setServerError] = useState<string | null>(null);

    const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginFormInputs>();

    const onSubmit = async (data: LoginFormInputs) => {
        setServerError(null);
        try {
            const response = await authApi.login(data);
            login(response.token, response.user_id);
            navigate("/profile");
        } catch (error: any) {
            setServerError(error.response?.data?.error || "Неверный email или пароль");
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 relative overflow-hidden font-sans">
            <div className="absolute inset-0 bg-gradient-to-br from-blue-600 via-blue-700 to-indigo-800 z-0" />
            <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-500/30 rounded-full blur-[100px] z-0" />

            <div className="w-full max-w-md bg-white rounded-3xl shadow-2xl p-8 z-10 relative m-4">
                <div className="flex flex-col items-center mb-8">
                    <div className="bg-blue-50 p-3 rounded-2xl mb-4">
                        <Plane className="h-8 w-8 text-blue-600 -rotate-45" />
                    </div>
                    <h1 className="text-2xl font-bold text-gray-900">С возвращением!</h1>
                    <p className="text-gray-500 text-sm mt-1">Введите данные для входа в Flyte</p>
                </div>

                {serverError && (
                    <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-xl border border-red-100 text-center">
                        {serverError}
                    </div>
                )}

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                    <div className="space-y-1">
                        <label className="text-sm font-medium text-gray-700 ml-1">Email</label>
                        <div className="relative">
                            <Mail className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
                            <input
                                {...register("email", { required: "Введите email" })}
                                type="email"
                                className={cn(
                                    "w-full h-12 pl-12 pr-4 rounded-xl bg-gray-50 border border-gray-200 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900",
                                    errors.email && "ring-2 ring-red-500 bg-red-50"
                                )}
                                placeholder="name@example.com"
                            />
                        </div>
                        {errors.email && <span className="text-xs text-red-500 ml-1">{errors.email.message}</span>}
                    </div>

                    <div className="space-y-1">
                        <label className="text-sm font-medium text-gray-700 ml-1">Пароль</label>
                        <div className="relative">
                            <Lock className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" />
                            <input
                                {...register("password", { required: "Введите пароль" })}
                                type="password"
                                className={cn(
                                    "w-full h-12 pl-12 pr-4 rounded-xl bg-gray-50 border border-gray-200 focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition text-gray-900",
                                    errors.password && "ring-2 ring-red-500 bg-red-50"
                                )}
                                placeholder="••••••••"
                            />
                        </div>
                        {errors.password && <span className="text-xs text-red-500 ml-1">{errors.password.message}</span>}
                    </div>

                    <button
                        disabled={isSubmitting}
                        type="submit"
                        className="w-full h-12 mt-2 bg-blue-600 text-white font-bold rounded-xl hover:bg-blue-700 active:scale-95 transition shadow-lg shadow-blue-500/30 flex items-center justify-center gap-2"
                    >
                        {isSubmitting ? "Вход..." : "Войти"}
                        {!isSubmitting && <ArrowRight className="h-5 w-5" />}
                    </button>
                </form>

                <div className="mt-8 text-center text-sm text-gray-500">
                    Нет аккаунта?{" "}
                    <Link to="/register" className="text-blue-600 font-bold hover:underline">
                        Зарегистрироваться
                    </Link>
                </div>
            </div>
        </div>
    );
};
