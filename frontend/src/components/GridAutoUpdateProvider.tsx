import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from 'react'
import { UpdateHeroesLayout } from '../../wailsjs/go/main/App'
import Toast from './Toast'
import type { ToastStatus } from './Toast'

interface GridAutoUpdateContextValue {
  scheduleGridUpdate: () => void
  cancelScheduledUpdate: () => void
  isAutoUpdating: boolean
}

const GridAutoUpdateContext = createContext<GridAutoUpdateContextValue | null>(null)

const AUTO_UPDATE_DEBOUNCE_MS = 500
const TOAST_SUCCESS_DURATION_MS = 5000

export function GridAutoUpdateProvider({ children }: { children: ReactNode }) {
  const [isAutoUpdating, setIsAutoUpdating] = useState(false)
  const [toast, setToast] = useState<{ status: ToastStatus; message: string } | null>(null)

  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const pendingRef = useRef(false)
  const updatingRef = useRef(false)
  const successTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const dismissToast = useCallback(() => {
    setToast(null)
    if (successTimerRef.current) {
      clearTimeout(successTimerRef.current)
      successTimerRef.current = null
    }
  }, [])

  const runUpdate = useCallback(async () => {
    updatingRef.current = true
    setIsAutoUpdating(true)
    dismissToast()
    setToast({ status: 'in-progress', message: 'Updating grids...' })

    try {
      await UpdateHeroesLayout()
      setToast({ status: 'success', message: 'Grids updated' })
      successTimerRef.current = setTimeout(() => {
        setToast(null)
        successTimerRef.current = null
      }, TOAST_SUCCESS_DURATION_MS)
    } catch (err) {
      setToast({ status: 'error', message: `Grid update failed: ${err}` })
    } finally {
      setIsAutoUpdating(false)
      updatingRef.current = false

      if (pendingRef.current) {
        pendingRef.current = false
        runUpdate()
      }
    }
  }, [dismissToast])

  const scheduleGridUpdate = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
    }

    if (updatingRef.current) {
      pendingRef.current = true
      return
    }

    timerRef.current = setTimeout(() => {
      timerRef.current = null
      runUpdate()
    }, AUTO_UPDATE_DEBOUNCE_MS)
  }, [runUpdate])

  const cancelScheduledUpdate = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
  }, [])

  return (
    <GridAutoUpdateContext.Provider value={{ scheduleGridUpdate, cancelScheduledUpdate, isAutoUpdating }}>
      {children}
      {toast && (
        <Toast status={toast.status} message={toast.message} onDismiss={dismissToast} />
      )}
    </GridAutoUpdateContext.Provider>
  )
}

export function useGridAutoUpdate(): GridAutoUpdateContextValue {
  const context = useContext(GridAutoUpdateContext)
  if (!context) {
    throw new Error('useGridAutoUpdate must be used within a GridAutoUpdateProvider')
  }
  return context
}
