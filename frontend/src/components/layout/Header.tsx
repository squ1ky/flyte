import { Link } from "react-router-dom";
import { Plane, User, LogIn } from "lucide-react";
import { useAuthStore } from "../../store/authStore";
import { cn } from "../../lib/utils";

interface HeaderProps {
    className?: string;
    transparent?: boolean;
}

export const Header = ({ className, transparent = false }: HeaderProps) => {
    const { isAuthenticated } = useAuthStore();

    return (
        <header
            className={cn(
                "absolute top-0 left-0 right-0 z-50 flex items-center justify-between px-6 py-4",
                !transparent && "bg-white shadow-sm relative",
                className
            )}
        >
            <Link to="/" className="flex items-center gap-2 hover:opacity-90 transition">
                <Plane
                    className={cn(
                        "h-8 w-8 -rotate-45",
                        transparent ? "text-white" : "text-blue-600"
                    )}
                    strokeWidth={2.5}
                />
                <span
                    className={cn(
                        "text-2xl font-bold tracking-tight",
                        transparent ? "text-white" : "text-gray-900"
                    )}
                >
                    Flyte
                </span>
            </Link>

            <nav className="flex items-center gap-4">
                {isAuthenticated ? (
                    <Link
                        to="/profile"
                        className={cn(
                            "flex items-center gap-2 px-4 py-2 rounded-full transition font-medium",
                            transparent
                                ? "bg-white/10 text-white backdrop-blur-md hover:bg-white/20"
                                : "bg-blue-50 text-blue-700 hover:bg-blue-100"
                        )}
                    >
                        <User className="h-5 w-5" />
                        <span>Профиль</span>
                    </Link>
                ) : (
                    <Link
                        to="/login"
                        className={cn(
                            "flex items-center gap-2 px-4 py-2 rounded-full transition font-medium",
                            transparent
                                ? "bg-white/10 text-white backdrop-blur-md hover:bg-white/20"
                                : "bg-gray-100 text-gray-900 hover:bg-gray-200"
                        )}
                    >
                        <LogIn className="h-5 w-5" />
                        <span>Войти</span>
                    </Link>
                )}
            </nav>
        </header>
    );
};
