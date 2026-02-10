import { Header } from "../../components/layout/Header";
import { useAuthStore } from "../../store/authStore";

export const ProfilePage = () => {
    const logout = useAuthStore((state: { logout: any; }) => state.logout);

    return (
        <div className="min-h-screen bg-gray-50 font-sans">
            <Header />

            <div className="pt-32 px-6 max-w-4xl mx-auto">
                <h1 className="text-3xl font-bold text-gray-900 mb-6">Личный кабинет</h1>
                <div className="bg-white rounded-2xl p-6 shadow-sm border border-gray-100">
                    <p className="text-gray-600 mb-4">Добро пожаловать в ваш профиль!</p>

                    <button
                        onClick={logout}
                        className="px-4 py-2 bg-red-50 text-red-600 rounded-lg font-medium hover:bg-red-100 transition"
                    >
                        Выйти
                    </button>
                </div>
            </div>
        </div>
    );
};
