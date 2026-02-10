import { Plane, User } from "lucide-react";
import { Link } from "react-router-dom";
import { cn } from "../../lib/utils";

export const Header = () => {
    return (
        <header className="absolute top-0 left-0 right-0 z-50 flex items-center justify-between px-6 py-4">
            <Link to="/" className="flex items-center gap-2 text-white hover:opacity-90 transition">
                <Plane className="h-8 w-8 -rotate-45" strokeWidth={2.5} />
                <span className="text-2xl font-bold tracking-tight">Flyte</span>
            </Link>

            <nav className="flex items-center gap-4">
                <Link
                    to="/login"
                    className={cn(
                        "flex items-center gap-2 px-4 py-2 rounded-full",
                        "bg-white/10 text-white backdrop-blur-md hover:bg-white/20 transition font-medium"
                    )}
                >
                    <User className="h-5 w-5" />
                    <span>Войти</span>
                </Link>
            </nav>
        </header>
    );
};
