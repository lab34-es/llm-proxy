import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import Box from '@mui/joy/Box'
import Sheet from '@mui/joy/Sheet'
import List from '@mui/joy/List'
import ListItem from '@mui/joy/ListItem'
import ListItemButton from '@mui/joy/ListItemButton'
import ListItemContent from '@mui/joy/ListItemContent'
import ListItemDecorator from '@mui/joy/ListItemDecorator'
import Typography from '@mui/joy/Typography'
import Divider from '@mui/joy/Divider'
import IconButton from '@mui/joy/IconButton'
import DnsIcon from '@mui/icons-material/Dns'
import VpnKeyIcon from '@mui/icons-material/VpnKey'
import BarChartIcon from '@mui/icons-material/BarChart'
import ShieldIcon from '@mui/icons-material/Shield'
import SmartToyIcon from '@mui/icons-material/SmartToy'
import LogoutIcon from '@mui/icons-material/Logout'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import LightModeIcon from '@mui/icons-material/LightMode'
import { useColorScheme } from '@mui/joy/styles'
import { useAuth } from '../context/AuthContext'

const NAV_ITEMS = [
  { path: 'providers', label: 'Providers', icon: <DnsIcon /> },
  { path: 'keys', label: 'API Keys', icon: <VpnKeyIcon /> },
  { path: 'usage', label: 'Usage', icon: <BarChartIcon /> },
  { path: 'guardrails', label: 'Guardrails', icon: <ShieldIcon /> },
  { path: 'playground', label: 'Playground', icon: <SmartToyIcon /> },
]

function ThemeToggle() {
  const { mode, setMode } = useColorScheme()
  return (
    <IconButton
      size="sm"
      variant="plain"
      color="neutral"
      onClick={() => setMode(mode === 'dark' ? 'light' : 'dark')}
      sx={{ ml: 'auto' }}
    >
      {mode === 'dark' ? <LightModeIcon /> : <DarkModeIcon />}
    </IconButton>
  )
}

export default function Layout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { logout } = useAuth()

  const currentPath = location.pathname.replace(/^\//, '').split('/')[0]

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      {/* Sidebar */}
      <Sheet
        sx={{
          width: 240,
          flexShrink: 0,
          display: 'flex',
          flexDirection: 'column',
          borderRight: '1px solid',
          borderColor: 'divider',
        }}
      >
        <Box sx={{ p: 2, display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography level="title-lg" sx={{ flexGrow: 1 }}>
            LLM Proxy
          </Typography>
          <ThemeToggle />
        </Box>
        <Divider />
        <List
          sx={{
            '--ListItem-radius': '8px',
            '--List-padding': '8px',
            '--List-gap': '4px',
            flexGrow: 1,
          }}
        >
          {NAV_ITEMS.map((item) => (
            <ListItem key={item.path}>
              <ListItemButton
                selected={currentPath === item.path}
                onClick={() => navigate(item.path)}
              >
                <ListItemDecorator>{item.icon}</ListItemDecorator>
                <ListItemContent>{item.label}</ListItemContent>
              </ListItemButton>
            </ListItem>
          ))}
        </List>
        <Divider />
        <List
          sx={{
            '--ListItem-radius': '8px',
            '--List-padding': '8px',
            '--List-gap': '4px',
          }}
        >
          <ListItem>
            <ListItemButton
              onClick={() => {
                logout()
                navigate('/login')
              }}
            >
              <ListItemDecorator>
                <LogoutIcon />
              </ListItemDecorator>
              <ListItemContent>Logout</ListItemContent>
            </ListItemButton>
          </ListItem>
        </List>
      </Sheet>

      {/* Main content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          overflow: 'auto',
        }}
      >
        <Outlet />
      </Box>
    </Box>
  )
}
