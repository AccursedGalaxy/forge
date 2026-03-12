import { Routes, Route } from 'react-router-dom'
import AppShell from './components/forge/AppShell'
import LandingPage from './pages/LandingPage'
import DashboardPage from './pages/DashboardPage'
import ContextPage from './pages/ContextPage'
import LogsPage from './pages/LogsPage'
import SettingsPage from './pages/SettingsPage'

function App() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/dashboard" element={<AppShell />}>
        <Route index element={<DashboardPage />} />
        <Route path="context" element={<ContextPage />} />
        <Route path="logs" element={<LogsPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>
    </Routes>
  )
}

export default App
