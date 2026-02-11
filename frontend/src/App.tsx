import { BrowserRouter, Routes, Route } from "react-router-dom";
import { HomePage } from "./pages/public/HomePage";
import { LoginPage } from "./pages/public/LoginPage";
import { RegisterPage } from "./pages/public/RegisterPage";
import { ProfilePage } from "./pages/private/ProfilePage";
import { ProtectedRoute } from "./features/auth/ProtectedRoute";
import {BookingPage} from "./pages/private/BookingPage.tsx";

function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />

                <Route element={<ProtectedRoute />}>
                    <Route path="/profile" element={<ProfilePage />} />
                    
                    <Route path="/flights/:flightId/book" element={<BookingPage />} />
                </Route>
            </Routes>
        </BrowserRouter>
    );
}

export default App;
