import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
    token: string | null;
    isAuthenticated: boolean;
    login: (token: string) => void;
    logout: () => void;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set: (arg0: { token: any; isAuthenticated: boolean; }) => void) => ({
            token: null,
            isAuthenticated: false,
            login: (token: any) => set({ token, isAuthenticated: true }),
            logout: () => {
                set({ token: null, isAuthenticated: false });
                localStorage.removeItem('auth-storage');
            },
        }),
        {
            name: 'auth-storage',
        }
    )
);
