import { Navigate, Route, Routes } from "react-router-dom";
import { AdminLayout } from "./layouts/AdminLayout";
import { AuthLayout } from "./layouts/AuthLayout";
import { ApplicationsPage } from "./pages/ApplicationsPage";
import { AuditLogsPage } from "./pages/AuditLogsPage";
import { DashboardPage } from "./pages/DashboardPage";
import { IdentitiesPage } from "./pages/IdentitiesPage";
import { LoginPage } from "./pages/LoginPage";
import { SecurityEventsPage } from "./pages/SecurityEventsPage";
import { SessionsPage } from "./pages/SessionsPage";
import { SettingsPage } from "./pages/SettingsPage";
import { getAdminSession } from "./store/authStore";

function RequireAdmin({ children }: { children: JSX.Element }) {
  if (!getAdminSession()) {
    return <Navigate to="/login" replace />;
  }
  return children;
}

export function App() {
  return (
    <Routes>
      <Route element={<AuthLayout />}>
        <Route path="/login" element={<LoginPage />} />
      </Route>
      <Route
        element={
          <RequireAdmin>
            <AdminLayout />
          </RequireAdmin>
        }
      >
        <Route path="/" element={<DashboardPage />} />
        <Route path="/applications" element={<ApplicationsPage />} />
        <Route path="/identities" element={<IdentitiesPage />} />
        <Route path="/sessions" element={<SessionsPage />} />
        <Route path="/audit-logs" element={<AuditLogsPage />} />
        <Route path="/security-events" element={<SecurityEventsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
