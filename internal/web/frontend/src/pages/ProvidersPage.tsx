import { useState, useEffect, useCallback, type FormEvent } from 'react'
import Box from '@mui/joy/Box'
import Typography from '@mui/joy/Typography'
import Table from '@mui/joy/Table'
import Sheet from '@mui/joy/Sheet'
import Button from '@mui/joy/Button'
import IconButton from '@mui/joy/IconButton'
import Input from '@mui/joy/Input'
import FormControl from '@mui/joy/FormControl'
import FormLabel from '@mui/joy/FormLabel'
import DismissibleAlert from '../components/DismissibleAlert'
import Card from '@mui/joy/Card'
import CardContent from '@mui/joy/CardContent'
import Modal from '@mui/joy/Modal'
import ModalDialog from '@mui/joy/ModalDialog'
import ModalClose from '@mui/joy/ModalClose'
import Stack from '@mui/joy/Stack'
import DeleteIcon from '@mui/icons-material/Delete'
import AddIcon from '@mui/icons-material/Add'
import { listProviders, createProvider, deleteProvider, type Provider } from '../api/client'

export default function ProvidersPage() {
  const [providers, setProviders] = useState<Provider[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [flash, setFlash] = useState('')
  const [modalOpen, setModalOpen] = useState(false)

  const [name, setName] = useState('')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const data = await listProviders()
      setProviders(data || [])
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load providers')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    if (!name || !baseUrl || !apiKey) {
      setError('All fields are required.')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      await createProvider({ name, base_url: baseUrl, api_key: apiKey })
      setFlash('Provider created successfully.')
      setName('')
      setBaseUrl('')
      setApiKey('')
      setModalOpen(false)
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to create provider')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('Delete this provider? All associated API keys will also be deleted.')) return
    try {
      await deleteProvider(id)
      setFlash('Provider deleted.')
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to delete provider')
    }
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography level="h3">Providers</Typography>
        <Button startDecorator={<AddIcon />} onClick={() => setModalOpen(true)}>
          Add Provider
        </Button>
      </Box>

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
                  <th>Base URL</th>
                  <th>Created</th>
                  <th style={{ width: 60 }}></th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={5}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>Loading...</Typography></td></tr>
                ) : providers.length === 0 ? (
                  <tr><td colSpan={5}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>No providers configured.</Typography></td></tr>
                ) : (
                  providers.map((p) => (
                    <tr key={p.id}>
                      <td><Typography level="body-xs" fontFamily="monospace">{p.id.slice(0, 8)}</Typography></td>
                      <td>{p.name}</td>
                      <td><Typography level="body-xs" fontFamily="monospace">{p.base_url}</Typography></td>
                      <td><Typography level="body-xs">{new Date(p.created_at).toLocaleString()}</Typography></td>
                      <td>
                        <IconButton size="sm" color="danger" variant="plain" onClick={() => handleDelete(p.id)}>
                          <DeleteIcon />
                        </IconButton>
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
          <Typography level="h4">Add Provider</Typography>
          <form onSubmit={handleCreate}>
            <Stack spacing={2} sx={{ mt: 1 }}>
              <FormControl required>
                <FormLabel>Name</FormLabel>
                <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. OpenAI" />
              </FormControl>
              <FormControl required>
                <FormLabel>Base URL</FormLabel>
                <Input value={baseUrl} onChange={(e) => setBaseUrl(e.target.value)} placeholder="https://api.openai.com" />
              </FormControl>
              <FormControl required>
                <FormLabel>API Key</FormLabel>
                <Input type="password" value={apiKey} onChange={(e) => setApiKey(e.target.value)} placeholder="sk-..." />
              </FormControl>
              <Button type="submit" loading={submitting}>Create Provider</Button>
            </Stack>
          </form>
        </ModalDialog>
      </Modal>
    </Box>
  )
}
