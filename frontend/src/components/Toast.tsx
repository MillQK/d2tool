import { XIcon } from './Icons'

export type ToastStatus = 'in-progress' | 'success' | 'error'

interface ToastProps {
  status: ToastStatus
  message: string
  onDismiss: () => void
}

function Toast({ status, message, onDismiss }: ToastProps) {
  return (
    <div className={`toast toast-${status}`}>
      <span className="toast-message">{message}</span>
      <button className="toast-dismiss" onClick={onDismiss}>
        <XIcon />
      </button>
    </div>
  )
}

export default Toast
