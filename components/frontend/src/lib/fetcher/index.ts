/**
 * TODO: Better error handling
 */
import { signOut } from 'next-auth/react'
import { createQueryString } from '../search'

export const getFetcher = (url: string) => {
  return async () => {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    return responseHandler(response)
  }
}

export const getPaginatedFetcher = (url: string, params?: {}) => {
  return async () => {
    const response = await fetch(url + createQueryString(params), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    return responseHandler(response)
  }
}

export const postFetcher = (url: string) => {
  return async (body: any) => {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    })

    return responseHandler(response)
  }
}

export const patchFetcher = (url: string) => {
  return async (body: any) => {
    const response = await fetch(url, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    })

    return responseHandler(response)
  }
}

export const postFormDataFetcher = (url: string) => {
  return async (body: any) => {
    // Convert object to FormData
    const formData = new FormData()

    for (const [key, value] of Object.entries(body)) {
      if (value !== null && value !== undefined) {
        // Handle File objects specially
        if (value instanceof File) {
          formData.append(key, value)
        } else {
          // Convert other values to string
          formData.append(key, String(value))
        }
      }
    }

    const response = await fetch(url, {
      method: 'POST',
      // Don't set Content-Type header - browser will set it with proper boundary for FormData
      body: formData
    })

    return responseHandler(response)
  }
}

export const patchFormDataFetcher = (url: string) => {
  return async (body: any) => {
    // Convert object to FormData
    const formData = new FormData()

    for (const [key, value] of Object.entries(body)) {
      if (value !== null && value !== undefined) {
        // Handle File objects specially
        if (value instanceof File) {
          formData.append(key, value)
        } else {
          // Convert other values to string
          formData.append(key, String(value))
        }
      }
    }

    const response = await fetch(url, {
      method: 'PATCH',
      // Don't set Content-Type header - browser will set it with proper boundary for FormData
      body: formData
    })

    return responseHandler(response)
  }
}

export const deleteFetcher = (url: string) => {
  return async ({ id }: { id: string }) => {
    const response = await fetch(`${url}/${id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    return responseHandler(response)
  }
}

export const downloadFetcher = (url: string) => {
  return async (): Promise<void> => {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    if (!response.ok) {
      throw new Error(`Download failed: ${response.statusText}`)
    }

    // The API returns a redirect to the download URL
    // For a download endpoint, we want to trigger the browser's download
    const downloadUrl = response.url

    // Create a temporary link element to trigger download
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = '' // Let the server determine the filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }
}

export const serverFetcher = async <T = void>(action: () => Promise<T>) => {
  try {
    return await action()
  } catch (error) {
    // Only log errors when not in test environment
    if (process.env.NODE_ENV !== 'test') {
      console.error('Server Fetcher Error', error)
    }
    return null
  }
}

const responseHandler = async (response: Response) => {
  if (!response.ok) {
    if (response.status === 401) {
      signOut({ callbackUrl: '/login' })
      return
    }

    const errorMessage = await response.json()
    throw new Error(errorMessage.message)
  }

  return await response.json()
}
