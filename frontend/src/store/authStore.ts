import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
    token: string | null;
    userId: number | null;
    isAuthenticated: boolean;
    login: (token: string, userId: number) => void;
    logout: () => void;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set) => ({
            token: null,
            userId: null,
            isAuthenticated: false,

            login: (token, userId) => {
                set({ token, userId, isAuthenticated: true });
            },

            logout: () => {
                set({ token: null, userId: null, isAuthenticated: false });
                localStorage.removeItem('auth-storage');
            },
        }),
        {
            name: 'auth-storage',
        }
    )
);
