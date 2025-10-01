import { useQueryClient } from '@tanstack/react-query'
import { useCallback } from 'react'

export interface RetryConfig {
  maxRetries?: number
  baseDelay?: number
  retryDelays?: number[]
  shouldRetry?: (error: Error) => boolean
}

export interface InvalidationOptions {
  queryKeys: string[][]
  onSuccess?: () => void
}

/**
 * Custom hook for handling query invalidation with retry logic
 */
export function useRetryInvalidation() {
  const queryClient = useQueryClient()

  const isCriticalFailure = useCallback((error: Error): boolean => {
    return (
      error.message.includes('network') || error.message.includes('timeout')
    )
  }, [])

  const invalidateWithRetry = useCallback(
    async (
      { queryKeys, onSuccess }: InvalidationOptions,
      config: RetryConfig = {}
    ): Promise<void> => {
      const {
        maxRetries = 3,
        retryDelays = [1000, 2000, 4000], // 1s, 2s, 4s exponential backoff
        shouldRetry = isCriticalFailure
      } = config

      try {
        // Attempt initial invalidation
        await Promise.all(
          queryKeys.map((queryKey) =>
            queryClient.invalidateQueries({ queryKey })
          )
        )
        onSuccess?.()
      } catch (error) {
        const err = error as Error
        console.error('Query invalidation failed:', err)

        // Only retry for critical failures
        if (!shouldRetry(err)) {
          console.warn('Non-critical invalidation failure, skipping retry')
          onSuccess?.()
          return
        }

        // Retry with exponential backoff
        let lastError = err
        for (let attempt = 0; attempt < maxRetries; attempt++) {
          const delay =
            retryDelays[attempt] || retryDelays[retryDelays.length - 1]

          try {
            await new Promise((resolve) => setTimeout(resolve, delay))
            await Promise.all(
              queryKeys.map((queryKey) =>
                queryClient.invalidateQueries({ queryKey })
              )
            )
            // Success - break retry loop
            console.log(
              `Query invalidation succeeded after ${attempt + 1} retries`
            )
            break
          } catch (retryError) {
            lastError = retryError as Error
            console.warn(
              `Query invalidation retry ${attempt + 1}/${maxRetries} failed after ${delay}ms:`,
              retryError
            )
          }
        }

        // Always call onSuccess even if retries failed
        // The UI should continue to function
        onSuccess?.()
      }
    },
    [queryClient, isCriticalFailure]
  )

  const simpleInvalidateWithRetry = useCallback(
    async (
      queryKeys: string[][],
      onSuccess?: () => void,
      maxRetries: number = 1
    ): Promise<void> => {
      try {
        await Promise.all(
          queryKeys.map((queryKey) =>
            queryClient.invalidateQueries({ queryKey })
          )
        )
        onSuccess?.()
      } catch (error) {
        console.error('Simple query invalidation failed:', error)

        // Single retry for non-critical operations
        if (maxRetries > 0) {
          try {
            await new Promise((resolve) => setTimeout(resolve, 2000)) // 2s delay
            await Promise.all(
              queryKeys.map((queryKey) =>
                queryClient.invalidateQueries({ queryKey })
              )
            )
          } catch (retryError) {
            console.warn('Simple query invalidation retry failed:', retryError)
          }
        }

        onSuccess?.()
      }
    },
    [queryClient]
  )

  return {
    invalidateWithRetry,
    simpleInvalidateWithRetry
  }
}
