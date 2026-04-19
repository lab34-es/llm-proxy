import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './context/AuthContext'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import ProvidersPage from './pages/ProvidersPage'
import KeysPage from './pages/KeysPage'
import UsagePage from './pages/UsagePage'
import GuardrailsPage from './pages/GuardrailsPage'
import PlaygroundPage from './pages/PlaygroundPage'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuth()
  if (!token) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="providers" replace />} />
        <Route path="providers" element={<ProvidersPage />} />
        <Route path="keys" element={<KeysPage />} />
        <Route path="usage" element={<UsagePage />} />
        <Route path="guardrails" element={<GuardrailsPage />} />
        <Route path="playground" element={<PlaygroundPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  )
}
