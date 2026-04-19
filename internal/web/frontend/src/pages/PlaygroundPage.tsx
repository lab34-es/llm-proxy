import { useState, useRef, useEffect, useCallback } from 'react'
import Box from '@mui/joy/Box'
import Typography from '@mui/joy/Typography'
import Input from '@mui/joy/Input'
import Button from '@mui/joy/Button'
import FormControl from '@mui/joy/FormControl'
import FormLabel from '@mui/joy/FormLabel'
import Textarea from '@mui/joy/Textarea'
import Card from '@mui/joy/Card'
import CardContent from '@mui/joy/CardContent'
import Sheet from '@mui/joy/Sheet'
import Stack from '@mui/joy/Stack'
import DismissibleAlert from '../components/DismissibleAlert'
import SendIcon from '@mui/icons-material/Send'
import DeleteIcon from '@mui/icons-material/Delete'

interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
}

export default function PlaygroundPage() {
  const [apiKey, setApiKey] = useState('')
  const [model, setModel] = useState('')
  const [systemPrompt, setSystemPrompt] = useState('')
  const [input, setInput] = useState('')
  const [conversation, setConversation] = useState<Message[]>([])
  const [streaming, setStreaming] = useState(false)
  const [error, setError] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const abortRef = useRef<AbortController | null>(null)

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [])

  useEffect(() => { scrollToBottom() }, [conversation, scrollToBottom])

  async function sendMessage() {
    const text = input.trim()
    if (!apiKey.trim()) {
      setError('Please enter an API key.')
      return
    }
    if (!model.trim()) {
      setError('Please enter a model name.')
      return
    }
    if (!text) return

    setError('')
    const userMsg: Message = { role: 'user', content: text }
    const newConvo = [...conversation, userMsg]
    setConversation(newConvo)
    setInput('')

    const messages: Message[] = []
    if (systemPrompt.trim()) {
      messages.push({ role: 'system', content: systemPrompt.trim() })
    }
    messages.push(...newConvo)

    setStreaming(true)
    const assistantMsg: Message = { role: 'assistant', content: '' }
    setConversation([...newConvo, assistantMsg])

    const controller = new AbortController()
    abortRef.current = controller

    try {
      const resp = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey.trim()}`,
        },
        body: JSON.stringify({
          model: model.trim(),
          messages,
          stream: true,
        }),
        signal: controller.signal,
      })

      if (!resp.ok) {
        const errText = await resp.text()
        setConversation((prev) => {
          const updated = [...prev]
          updated[updated.length - 1] = { role: 'assistant', content: `Error: ${errText}` }
          return updated
        })
        setStreaming(false)
        return
      }

      const reader = resp.body!.getReader()
      const decoder = new TextDecoder()
      let assistantText = ''
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice(6).trim()
          if (data === '[DONE]') continue

          try {
            const parsed = JSON.parse(data)
            const delta = parsed.choices?.[0]?.delta?.content
            if (delta) {
              assistantText += delta
              setConversation((prev) => {
                const updated = [...prev]
                updated[updated.length - 1] = { role: 'assistant', content: assistantText }
                return updated
              })
            }
          } catch {
            // skip malformed chunks
          }
        }
      }
    } catch (err: unknown) {
      if (err instanceof Error && err.name !== 'AbortError') {
        setConversation((prev) => {
          const updated = [...prev]
          updated[updated.length - 1] = { role: 'assistant', content: `Error: ${err.message}` }
          return updated
        })
      }
    } finally {
      setStreaming(false)
      abortRef.current = null
    }
  }

  function clearConversation() {
    if (abortRef.current) abortRef.current.abort()
    setConversation([])
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey && !streaming) {
      e.preventDefault()
      sendMessage()
    }
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 48px)' }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography level="h3">Playground</Typography>
        <Button
          size="sm"
          variant="plain"
          color="danger"
          startDecorator={<DeleteIcon />}
          onClick={clearConversation}
        >
          Clear
        </Button>
      </Box>

      {error && <DismissibleAlert color="danger" sx={{ mb: 2 }} onClose={() => setError('')}>{error}</DismissibleAlert>}

      {/* Config */}
      <Card variant="outlined" sx={{ mb: 2 }}>
        <CardContent>
          <Stack direction="row" spacing={2} flexWrap="wrap">
            <FormControl size="sm" sx={{ flex: 1, minWidth: 200 }}>
              <FormLabel>API Key</FormLabel>
              <Input
                type="password"
                size="sm"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="llmp-..."
              />
            </FormControl>
            <FormControl size="sm" sx={{ flex: 1, minWidth: 200 }}>
              <FormLabel>Model</FormLabel>
              <Input
                size="sm"
                value={model}
                onChange={(e) => setModel(e.target.value)}
                placeholder="e.g. gpt-4"
              />
            </FormControl>
          </Stack>
          <FormControl size="sm" sx={{ mt: 1 }}>
            <FormLabel>System Prompt</FormLabel>
            <Textarea
              size="sm"
              minRows={1}
              maxRows={3}
              value={systemPrompt}
              onChange={(e) => setSystemPrompt(e.target.value)}
              placeholder="Optional system prompt..."
            />
          </FormControl>
        </CardContent>
      </Card>

      {/* Messages */}
      <Sheet
        variant="outlined"
        sx={{
          flex: 1,
          overflow: 'auto',
          borderRadius: 'sm',
          p: 2,
          mb: 2,
          display: 'flex',
          flexDirection: 'column',
          gap: 1.5,
        }}
      >
        {conversation.length === 0 ? (
          <Typography level="body-sm" sx={{ color: 'text.tertiary', textAlign: 'center', mt: 4 }}>
            Send a message to start a conversation.
          </Typography>
        ) : (
          conversation.map((msg, i) => (
            <Box
              key={i}
              sx={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: msg.role === 'user' ? 'flex-end' : 'flex-start',
              }}
            >
              <Card
                size="sm"
                variant={msg.role === 'user' ? 'solid' : 'soft'}
                color={msg.role === 'user' ? 'primary' : 'neutral'}
                sx={{ maxWidth: '80%' }}
              >
                <CardContent>
                  <Typography
                    level="body-sm"
                    sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}
                  >
                    {msg.content || (streaming && i === conversation.length - 1 ? '...' : '')}
                  </Typography>
                </CardContent>
              </Card>
              <Typography level="body-xs" sx={{ mt: 0.5, color: 'text.tertiary' }}>
                {msg.role === 'user' ? 'You' : 'Assistant'}
              </Typography>
            </Box>
          ))
        )}
        <div ref={messagesEndRef} />
      </Sheet>

      {/* Input */}
      <Box sx={{ display: 'flex', gap: 1 }}>
        <Input
          sx={{ flex: 1 }}
          placeholder="Type a message..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={streaming}
        />
        <Button
          onClick={sendMessage}
          disabled={streaming}
          loading={streaming}
          startDecorator={<SendIcon />}
        >
          Send
        </Button>
      </Box>
    </Box>
  )
}
