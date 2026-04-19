import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { CssVarsProvider } from '@mui/joy/styles'
import CssBaseline from '@mui/joy/CssBaseline'
import App from './App'
import { AuthProvider } from './context/AuthContext'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter basename="/dashboard">
      <CssVarsProvider defaultMode="system">
        <CssBaseline />
        <AuthProvider>
          <App />
        </AuthProvider>
      </CssVarsProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
