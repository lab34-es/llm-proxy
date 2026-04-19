import Alert from '@mui/joy/Alert'
import IconButton from '@mui/joy/IconButton'
import CloseIcon from '@mui/icons-material/Close'
import type { ColorPaletteProp } from '@mui/joy/styles'
import type { ReactNode } from 'react'

interface Props {
  color: ColorPaletteProp
  children: ReactNode
  onClose: () => void
  endDecorator?: ReactNode
  sx?: Record<string, unknown>
}

export default function DismissibleAlert({ color, children, onClose, endDecorator, sx }: Props) {
  return (
    <Alert
      color={color}
      sx={sx}
      endDecorator={
        <>
          {endDecorator}
          <IconButton size="sm" color={color} variant="plain" onClick={onClose}>
            <CloseIcon />
          </IconButton>
        </>
      }
    >
      {children}
    </Alert>
  )
}
