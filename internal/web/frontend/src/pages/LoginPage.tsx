import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import Box from '@mui/joy/Box'
import Card from '@mui/joy/Card'
import CardContent from '@mui/joy/CardContent'
import Typography from '@mui/joy/Typography'
import FormControl from '@mui/joy/FormControl'
import FormLabel from '@mui/joy/FormLabel'
import Input from '@mui/joy/Input'
import Button from '@mui/joy/Button'
import Alert from '@mui/joy/Alert'
import { useAuth } from '../context/AuthContext'

export default function LoginPage() {
  const [token, setToken] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!token.trim()) {
      setError('Admin token is required.')
      return
    }

    setLoading(true)
    setError('')

    try {
      // Validate token by making a test API call.
      const resp = await fetch('/admin/providers', {
        headers: { Authorization: `Bearer ${token.trim()}` },
      })
      if (resp.status === 401) {
        setError('Invalid admin token.')
        setLoading(false)
        return
      }
      if (!resp.ok) {
        setError('Failed to verify token.')
        setLoading(false)
        return
      }
      login(token.trim())
      navigate('/', { replace: true })
    } catch {
      setError('Failed to connect to the server.')
      setLoading(false)
    }
  }

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
      }}
    >
      <Card variant="outlined" sx={{ width: 380 }}>
        <CardContent>
          <Typography level="h3" sx={{ mb: 2 }}>
            LLM Proxy
          </Typography>
          <Typography level="body-sm" sx={{ mb: 3 }}>
            Enter your admin token to access the dashboard.
          </Typography>
          {error && (
            <Alert color="danger" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          <form onSubmit={handleSubmit}>
            <FormControl sx={{ mb: 2 }}>
              <FormLabel>Admin Token</FormLabel>
              <Input
                type="password"
                placeholder="Enter admin token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                autoFocus
              />
            </FormControl>
            <Button type="submit" fullWidth loading={loading}>
              Sign In
            </Button>
          </form>
        </CardContent>
      </Card>
    </Box>
  )
}
