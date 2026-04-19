import { useState, useEffect, useCallback } from 'react'
import Box from '@mui/joy/Box'
import Typography from '@mui/joy/Typography'
import Table from '@mui/joy/Table'
import Sheet from '@mui/joy/Sheet'
import Select from '@mui/joy/Select'
import Option from '@mui/joy/Option'
import Input from '@mui/joy/Input'
import FormControl from '@mui/joy/FormControl'
import FormLabel from '@mui/joy/FormLabel'
import Button from '@mui/joy/Button'
import DismissibleAlert from '../components/DismissibleAlert'
import Card from '@mui/joy/Card'
import CardContent from '@mui/joy/CardContent'
import Stack from '@mui/joy/Stack'
import Grid from '@mui/joy/Grid'
import { queryUsage, listKeys, listProviders, type UsageRecord, type APIKey, type Provider } from '../api/client'

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <Card variant="soft" size="sm">
      <CardContent>
        <Typography level="body-xs" textTransform="uppercase" fontWeight="bold">
          {label}
        </Typography>
        <Typography level="h3">{typeof value === 'number' ? value.toLocaleString() : value}</Typography>
      </CardContent>
    </Card>
  )
}

export default function UsagePage() {
  const [records, setRecords] = useState<UsageRecord[]>([])
  const [total, setTotal] = useState(0)
  const [keys, setKeys] = useState<APIKey[]>([])
  const [providers, setProviders] = useState<Provider[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Filters
  const [filterKeyId, setFilterKeyId] = useState('')
  const [filterProviderId, setFilterProviderId] = useState('')
  const [filterStart, setFilterStart] = useState('')
  const [filterEnd, setFilterEnd] = useState('')
  const [page, setPage] = useState(1)
  const perPage = 50

  const load = useCallback(async () => {
    try {
      setLoading(true)
      const params: Record<string, string> = {
        limit: String(perPage),
        offset: String((page - 1) * perPage),
      }
      if (filterKeyId) params.api_key_id = filterKeyId
      if (filterProviderId) params.provider_id = filterProviderId
      if (filterStart) params.start = new Date(filterStart).toISOString()
      if (filterEnd) params.end = new Date(filterEnd).toISOString()

      const [result, k, p] = await Promise.all([
        queryUsage(params),
        listKeys(),
        listProviders(),
      ])
      setRecords(result.records || [])
      setTotal(result.total)
      setKeys(k || [])
      setProviders(p || [])
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load usage')
    } finally {
      setLoading(false)
    }
  }, [page, filterKeyId, filterProviderId, filterStart, filterEnd])

  useEffect(() => { load() }, [load])

  const totalPages = Math.max(1, Math.ceil(total / perPage))

  // Compute page-level stats
  const promptTotal = records.reduce((s, r) => s + r.prompt_tokens, 0)
  const completionTotal = records.reduce((s, r) => s + r.completion_tokens, 0)
  const allTotal = records.reduce((s, r) => s + r.total_tokens, 0)

  function keyName(id: string): string {
    return keys.find((k) => k.id === id)?.name || id.slice(0, 8)
  }

  function providerName(id: string): string {
    return providers.find((p) => p.id === id)?.name || id.slice(0, 8)
  }

  function handleFilter() {
    setPage(1)
    load()
  }

  function clearFilters() {
    setFilterKeyId('')
    setFilterProviderId('')
    setFilterStart('')
    setFilterEnd('')
    setPage(1)
  }

  return (
    <Box>
      <Typography level="h3" sx={{ mb: 2 }}>Usage</Typography>

      {error && <DismissibleAlert color="danger" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</DismissibleAlert>}

      {/* Filters */}
      <Card variant="outlined" sx={{ mb: 2 }}>
        <CardContent>
          <Stack direction="row" spacing={2} flexWrap="wrap" alignItems="flex-end">
            <FormControl size="sm">
              <FormLabel>API Key</FormLabel>
              <Select
                size="sm"
                value={filterKeyId}
                onChange={(_, v) => setFilterKeyId(v || '')}
                placeholder="All keys"
                sx={{ minWidth: 150 }}
              >
                <Option value="">All keys</Option>
                {keys.map((k) => (
                  <Option key={k.id} value={k.id}>{k.name}</Option>
                ))}
              </Select>
            </FormControl>
            <FormControl size="sm">
              <FormLabel>Provider</FormLabel>
              <Select
                size="sm"
                value={filterProviderId}
                onChange={(_, v) => setFilterProviderId(v || '')}
                placeholder="All providers"
                sx={{ minWidth: 150 }}
              >
                <Option value="">All providers</Option>
                {providers.map((p) => (
                  <Option key={p.id} value={p.id}>{p.name}</Option>
                ))}
              </Select>
            </FormControl>
            <FormControl size="sm">
              <FormLabel>Start</FormLabel>
              <Input
                size="sm"
                type="datetime-local"
                value={filterStart}
                onChange={(e) => setFilterStart(e.target.value)}
              />
            </FormControl>
            <FormControl size="sm">
              <FormLabel>End</FormLabel>
              <Input
                size="sm"
                type="datetime-local"
                value={filterEnd}
                onChange={(e) => setFilterEnd(e.target.value)}
              />
            </FormControl>
            <Button size="sm" onClick={handleFilter}>Apply</Button>
            <Button size="sm" variant="plain" color="neutral" onClick={clearFilters}>Clear</Button>
          </Stack>
        </CardContent>
      </Card>

      {/* Stats */}
      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid xs={6} md={3}><StatCard label="Total Records" value={total} /></Grid>
        <Grid xs={6} md={3}><StatCard label="Prompt Tokens" value={promptTotal} /></Grid>
        <Grid xs={6} md={3}><StatCard label="Completion Tokens" value={completionTotal} /></Grid>
        <Grid xs={6} md={3}><StatCard label="Total Tokens" value={allTotal} /></Grid>
      </Grid>

      {/* Table */}
      <Card variant="outlined">
        <CardContent sx={{ p: 0 }}>
          <Sheet sx={{ overflow: 'auto' }}>
            <Table stickyHeader size="sm">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Key</th>
                  <th>Provider</th>
                  <th>Model</th>
                  <th>Prompt</th>
                  <th>Completion</th>
                  <th>Total</th>
                  <th>Time</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr><td colSpan={8}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>Loading...</Typography></td></tr>
                ) : records.length === 0 ? (
                  <tr><td colSpan={8}><Typography level="body-sm" sx={{ textAlign: 'center', py: 2 }}>No usage records.</Typography></td></tr>
                ) : (
                  records.map((r) => (
                    <tr key={r.id}>
                      <td><Typography level="body-xs" fontFamily="monospace">{r.id.slice(0, 8)}</Typography></td>
                      <td><Typography level="body-xs">{keyName(r.api_key_id)}</Typography></td>
                      <td><Typography level="body-xs">{providerName(r.provider_id)}</Typography></td>
                      <td><Typography level="body-xs" fontFamily="monospace">{r.model}</Typography></td>
                      <td>{r.prompt_tokens.toLocaleString()}</td>
                      <td>{r.completion_tokens.toLocaleString()}</td>
                      <td><strong>{r.total_tokens.toLocaleString()}</strong></td>
                      <td><Typography level="body-xs">{new Date(r.created_at).toLocaleString()}</Typography></td>
                    </tr>
                  ))
                )}
              </tbody>
            </Table>
          </Sheet>
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <Box sx={{ display: 'flex', justifyContent: 'center', gap: 1, mt: 2 }}>
          <Button size="sm" variant="plain" disabled={page <= 1} onClick={() => setPage(page - 1)}>Previous</Button>
          <Typography level="body-sm" sx={{ display: 'flex', alignItems: 'center' }}>
            Page {page} of {totalPages}
          </Typography>
          <Button size="sm" variant="plain" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>Next</Button>
        </Box>
      )}
    </Box>
  )
}
