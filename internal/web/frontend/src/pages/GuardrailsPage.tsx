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
import Divider from '@mui/joy/Divider'
import DeleteIcon from '@mui/icons-material/Delete'
import AddIcon from '@mui/icons-material/Add'
import {
  listGuardrails,
  createGuardrail,
  deleteGuardrail,
  listGuardrailEvents,
  deleteGuardrailEvent,
  type Guardrail,
  type GuardrailEvent,
} from '../api/client'

export default function GuardrailsPage() {
  const [guardrails, setGuardrails] = useState<Guardrail[]>([])
  const [events, setEvents] = useState<GuardrailEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [flash, setFlash] = useState('')
  const [modalOpen, setModalOpen] = useState(false)

  const [pattern, setPattern] = useState('')
  const [mode, setMode] = useState('reject')
  const [replaceBy, setReplaceBy] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const [g, e] = await Promise.all([
        listGuardrails(),
        listGuardrailEvents({ limit: '50' }),
      ])
      setGuardrails(g || [])
      setEvents(e.records || [])
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    if (!pattern || !mode) {
      setError('Pattern and Mode are required.')
      return
    }
    setSubmitting(true)
    setError('')
    try {
      await createGuardrail({
        pattern,
        mode,
        replace_by: mode === 'replace' ? replaceBy : undefined,
      })
      setFlash('Guardrail created.')
      setPattern('')
      setMode('reject')
      setReplaceBy('')
      setModalOpen(false)
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to create guardrail')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDeleteGuardrail(id: string) {
    if (!confirm('Delete this guardrail rule?')) return
    try {
      await deleteGuardrail(id)
      setFlash('Guardrail deleted.')
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to delete guardrail')
    }
  }

  async function handleDeleteEvent(id: string) {
    try {
      await deleteGuardrailEvent(id)
      setFlash('Event deleted.')
      await load()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to delete event')
    }
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography level="h3">Guardrails</Typography>
        <Button startDecorator={<AddIcon />} onClick={() => setModalOpen(true)}>
          Add Rule
        </Button>
      </Box>

      {flash && <DismissibleAlert color="success" sx={{ mb: 2 }} onClose={() => setFlash('')}>{flash}</DismissibleAlert>}
      {error && <DismissibleAlert color="danger" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</DismissibleAlert>}

      {/* Guardrail Rules */}
      <Card variant="outlined" sx={{ mb: 3 }}>
        <CardContent sx={{ p: 0 }}>
          <Sheet sx={{ overflow: 'auto' }}>
            <Table stickyHeader>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Pattern</th>
                  <th>Mode</th>
                  <th>Replacement</th>
                  <th>Created</th>
                  <th style={{ width: 60 }}></th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={6}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>Loading...</Typography></td></tr>
                ) : guardrails.length === 0 ? (
                  <tr><td colSpan={6}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>No guardrail rules.</Typography></td></tr>
                ) : (
                  guardrails.map((g) => (
                    <tr key={g.id}>
                      <td><Typography level="body-xs" fontFamily="monospace">{g.id.slice(0, 8)}</Typography></td>
                      <td><Typography level="body-xs" fontFamily="monospace">{g.pattern}</Typography></td>
                      <td>
                        <Chip
                          size="sm"
                          color={g.mode === 'reject' ? 'danger' : 'warning'}
                          variant="soft"
                        >
                          {g.mode}
                        </Chip>
                      </td>
                      <td><Typography level="body-xs">{g.replace_by || '-'}</Typography></td>
                      <td><Typography level="body-xs">{new Date(g.created_at).toLocaleString()}</Typography></td>
                      <td>
                        <IconButton size="sm" color="danger" variant="plain" onClick={() => handleDeleteGuardrail(g.id)}>
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

      {/* Guardrail Events */}
      <Divider sx={{ my: 3 }} />
      <Typography level="h4" sx={{ mb: 2 }}>Recent Events</Typography>
      <Card variant="outlined">
        <CardContent sx={{ p: 0 }}>
          <Sheet sx={{ overflow: 'auto' }}>
            <Table stickyHeader size="sm">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Pattern</th>
                  <th>Mode</th>
                  <th>API Key</th>
                  <th>Input Text</th>
                  <th>Time</th>
                  <th style={{ width: 60 }}></th>
                </tr>
              </thead>
              <tbody>
                {events.length === 0 ? (
                  <tr><td colSpan={7}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>No events recorded.</Typography></td></tr>
                ) : (
                  events.map((ev) => (
                    <tr key={ev.id}>
                      <td><Typography level="body-xs" fontFamily="monospace">{ev.id.slice(0, 8)}</Typography></td>
                      <td><Typography level="body-xs" fontFamily="monospace">{ev.pattern}</Typography></td>
                      <td>
                        <Chip size="sm" color={ev.mode === 'reject' ? 'danger' : 'warning'} variant="soft">
                          {ev.mode}
                        </Chip>
                      </td>
                      <td><Typography level="body-xs" fontFamily="monospace">{ev.api_key_id.slice(0, 8)}</Typography></td>
                      <td>
                        <Typography level="body-xs" sx={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {ev.input_text}
                        </Typography>
                      </td>
                      <td><Typography level="body-xs">{new Date(ev.created_at).toLocaleString()}</Typography></td>
                      <td>
                        <IconButton size="sm" color="danger" variant="plain" onClick={() => handleDeleteEvent(ev.id)}>
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

      {/* Create Modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)}>
        <ModalDialog>
          <ModalClose />
          <Typography level="h4">Add Guardrail Rule</Typography>
          <form onSubmit={handleCreate}>
            <Stack spacing={2} sx={{ mt: 1 }}>
              <FormControl required>
                <FormLabel>Pattern (regex)</FormLabel>
                <Input
                  value={pattern}
                  onChange={(e) => setPattern(e.target.value)}
                  placeholder="e.g. \bpassword\b"
                  slotProps={{ input: { style: { fontFamily: 'monospace' } } }}
                />
              </FormControl>
              <FormControl required>
                <FormLabel>Mode</FormLabel>
                <Select value={mode} onChange={(_, v) => setMode(v || 'reject')}>
                  <Option value="reject">Reject</Option>
                  <Option value="replace">Replace</Option>
                </Select>
              </FormControl>
              {mode === 'replace' && (
                <FormControl>
                  <FormLabel>Replacement Text</FormLabel>
                  <Input value={replaceBy} onChange={(e) => setReplaceBy(e.target.value)} placeholder="e.g. [REDACTED]" />
                </FormControl>
              )}
              <Button type="submit" loading={submitting}>Create Rule</Button>
            </Stack>
          </form>
        </ModalDialog>
      </Modal>
    </Box>
  )
}
