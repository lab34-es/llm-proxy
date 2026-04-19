import { useState, useEffect, useCallback, type FormEvent } from 'react'
import Box from '@mui/joy/Box'
import Typography from '@mui/joy/Typography'
import Table from '@mui/joy/Table'
import Sheet from '@mui/joy/Sheet'
import Button from '@mui/joy/Button'
import IconButton from '@mui/joy/IconButton'
import Input from '@mui/joy/Input'
import Select from '@mui/joy/Select'
import Option from '@mui/joy/Option'
import FormControl from '@mui/joy/FormControl'
import FormLabel from '@mui/joy/FormLabel'
import DismissibleAlert from '../components/DismissibleAlert'
import Card from '@mui/joy/Card'
import CardContent from '@mui/joy/CardContent'
import Chip from '@mui/joy/Chip'
import Modal from '@mui/joy/Modal'
import ModalDialog from '@mui/joy/ModalDialog'
import ModalClose from '@mui/joy/ModalClose'
import Stack from '@mui/joy/Stack'
import DeleteIcon from '@mui/icons-material/Delete'
import AddIcon from '@mui/icons-material/Add'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import { listKeys, createKey, revokeKey, listProviders, type APIKey, type Provider } from '../api/client'

export default function KeysPage() {
  const [keys, setKeys] = useState<APIKey[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [flash, setFlash] = useState('')
  const [newKey, setNewKey] = useState('')
  const [modalOpen, setModalOpen] = useState(false)

  const [name, setName] = useState('')
  const [providerId, setProviderId] = useState('')
  const [rpm, setRpm] = useState('60')
  const [submitting, setSubmitting] = useState(false)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [k, p] = await Promise.all([listKeys(), listProviders()])
      setKeys(k || [])
      setProviders(p || [])
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    if (!name || !providerId) {
      setError('Name and Provider are required.')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      const result = await createKey({
        name,
        provider_id: providerId,
        rate_limit_rpm: parseInt(rpm) || 60,
      })
      setNewKey(result.key)
      setFlash('API key created. Copy it now - it will not be shown again.')
      setName('')
      setProviderId('')
      setRpm('60')
      setModalOpen(false)
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to create key')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleRevoke(id: string) {
    if (!confirm('Revoke this API key? It will no longer be usable.')) return
    try {
      await revokeKey(id)
      setFlash('Key revoked.')
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to revoke key')
    }
  }

  function providerName(id: string): string {
    return providers.find((p) => p.id === id)?.name || id.slice(0, 8)
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography level="h3">API Keys</Typography>
        <Button startDecorator={<AddIcon />} onClick={() => setModalOpen(true)}>
          Create Key
        </Button>
      </Box>

      {newKey && (
        <DismissibleAlert
          color="warning"
          sx={{ mb: 2 }}
          endDecorator={
            <IconButton
              size="sm"
              color="warning"
              variant="plain"
              onClick={() => { navigator.clipboard.writeText(newKey); setFlash('Key copied to clipboard.') }}
            >
              <ContentCopyIcon />
            </IconButton>
          }
          onClose={() => setNewKey('')}
        >
          <Box>
            <Typography level="body-sm" fontWeight="bold">New API Key (copy it now!):</Typography>
            <Typography level="body-xs" sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>{newKey}</Typography>
          </Box>
        </DismissibleAlert>
      )}

      {flash && <DismissibleAlert color="success" sx={{ mb: 2 }} onClose={() => setFlash('')}>{flash}</DismissibleAlert>}
      {error && <DismissibleAlert color="danger" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</DismissibleAlert>}

      <Card variant="outlined">
        <CardContent sx={{ p: 0 }}>
          <Sheet sx={{ overflow: 'auto' }}>
            <Table stickyHeader>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Name</th>
                  <th>Provider</th>
                  <th>RPM</th>
                  <th>Status</th>
                  <th>Created</th>
                  <th style={{ width: 60 }}></th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={7}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>Loading...</Typography></td></tr>
                ) : keys.length === 0 ? (
                  <tr><td colSpan={7}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>No API keys created.</Typography></td></tr>
                ) : (
                  keys.map((k) => (
                    <tr key={k.id}>
                      <td><Typography level="body-xs" fontFamily="monospace">{k.id.slice(0, 8)}</Typography></td>
                      <td>{k.name}</td>
                      <td>{providerName(k.provider_id)}</td>
                      <td>{k.rate_limit_rpm}</td>
                      <td>
                        {k.revoked_at ? (
                          <Chip size="sm" color="danger" variant="soft">Revoked</Chip>
                        ) : (
                          <Chip size="sm" color="success" variant="soft">Active</Chip>
                        )}
                      </td>
                      <td><Typography level="body-xs">{new Date(k.created_at).toLocaleString()}</Typography></td>
                      <td>
                        {!k.revoked_at && (
                          <IconButton size="sm" color="danger" variant="plain" onClick={() => handleRevoke(k.id)}>
                            <DeleteIcon />
                          </IconButton>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </Table>
          </Sheet>
        </CardContent>
      </Card>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)}>
        <ModalDialog>
          <ModalClose />
          <Typography level="h4">Create API Key</Typography>
          <form onSubmit={handleCreate}>
            <Stack spacing={2} sx={{ mt: 1 }}>
              <FormControl required>
                <FormLabel>Name</FormLabel>
                <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. my-app-key" />
              </FormControl>
              <FormControl required>
                <FormLabel>Provider</FormLabel>
                <Select
                  value={providerId}
                  onChange={(_, v) => setProviderId(v || '')}
                  placeholder="Select a provider"
                >
                  {providers.map((p) => (
                    <Option key={p.id} value={p.id}>{p.name}</Option>
                  ))}
                </Select>
              </FormControl>
              <FormControl>
                <FormLabel>Rate Limit (RPM)</FormLabel>
                <Input type="number" value={rpm} onChange={(e) => setRpm(e.target.value)} />
              </FormControl>
              <Button type="submit" loading={submitting}>Create Key</Button>
            </Stack>
          </form>
        </ModalDialog>
      </Modal>
    </Box>
  )
}
