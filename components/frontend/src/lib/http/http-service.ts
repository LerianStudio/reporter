import { createQueryString } from '@/lib/search'
import { HttpStatus } from './http-status'
import {
  ApiException,
  InternalServerErrorApiException,
  NotFoundApiException,
  ServiceUnavailableApiException,
  UnauthorizedApiException,
  UnprocessableEntityApiException
} from './api-exception'
import {
  forOwn,
  isNull,
  isUndefined,
  forEach,
  isArray,
  isObject,
  toString
} from 'lodash'

export interface FetchModuleOptions extends RequestInit {
  baseUrl?: URL | string
  search?: object
}

/**
 * HTTP service class to allow easy implementation of custom API repositories
 *
 * Code based from nestjs-fetch:
 * https://github.com/mikehall314/nestjs-fetch/blob/main/lib/fetch.service.ts
 */
export abstract class HttpService {
  protected async request<T>(request: Request): Promise<T> {
    try {
      this.onBeforeFetch(request)

      const response = await fetch(request)

      this.onAfterFetch(request, response)

      // Parse text/plain error responses
      if (response?.headers?.get('content-type')?.includes('text/plain')) {
        const message = await response.text()

        await this.catch(request, response, { message })

        if (response.status === HttpStatus.UNAUTHORIZED) {
          throw new UnauthorizedApiException(message)
        } else if (response.status === HttpStatus.NOT_FOUND) {
          throw new NotFoundApiException(message)
        } else if (response.status === HttpStatus.UNPROCESSABLE_ENTITY) {
          throw new UnprocessableEntityApiException(message)
        } else if (response.status === HttpStatus.INTERNAL_SERVER_ERROR) {
          throw new InternalServerErrorApiException(message)
        }

        throw new ServiceUnavailableApiException(message)
      }

      // Parse application/json error responses
      // NodeJS native fetch does not throw for logic errors
      if (!response.ok) {
        const error = await response.json()

        await this.catch(request, response, error)

        if (response.status === HttpStatus.UNAUTHORIZED) {
          throw new UnauthorizedApiException(error)
        } else if (response.status === HttpStatus.NOT_FOUND) {
          throw new NotFoundApiException(error)
        } else if (response.status === HttpStatus.UNPROCESSABLE_ENTITY) {
          throw new UnprocessableEntityApiException(error)
        } else if (response.status === HttpStatus.INTERNAL_SERVER_ERROR) {
          throw new InternalServerErrorApiException(error)
        }

        throw new ServiceUnavailableApiException(error)
      }

      // Handle 204 Success No Content response
      if (response.status === HttpStatus.NO_CONTENT) {
        return {} as T
      }

      return await response.json()
    } catch (error: any) {
      if (error instanceof ApiException) {
        throw error
      }

      throw new ServiceUnavailableApiException(error)
    }
  }

  private async createRequest(
    url: URL | string,
    options: FetchModuleOptions
  ): Promise<Request> {
    const { baseUrl, search, ...init } = {
      ...(await this.createDefaults()),
      ...options
    }

    return new Request(new URL(url + createQueryString(search), baseUrl), {
      ...init,
      ...options,
      headers: {
        ...init?.headers,
        ...options?.headers
      }
    })
  }

  protected async createDefaults() {
    return {}
  }

  /**
   * Event triggered before the request is sent
   * @param request The request to be sent
   */
  protected onBeforeFetch(request: Request) {}

  /**
   * Event triggered after the request is sent, but before response JSON parsing
   * @param request The request that was sent
   * @param response The raw response received from the server
   */
  protected onAfterFetch(request: Request, response: Response) {}

  /**
   * Catch function to handle errors from the native fetch API
   * @param request The request that was sent
   * @param response The raw response received from the server
   * @param error Parsed error response from the server
   */
  protected async catch(request: Request, response: Response, error: any) {
    console.error('Request error', { response, error })
  }

  async get<T>(
    url: URL | string,
    options: FetchModuleOptions = {}
  ): Promise<T> {
    const request = await this.createRequest(url, { ...options, method: 'GET' })
    return this.request<T>(request)
  }

  async head(
    url: URL | string,
    options: FetchModuleOptions = {}
  ): Promise<Response> {
    const request = await this.createRequest(url, {
      ...options,
      method: 'HEAD'
    })
    return this.request(request)
  }

  async delete(
    url: URL | string,
    options: FetchModuleOptions = {}
  ): Promise<Response> {
    const request = await this.createRequest(url, {
      ...options,
      method: 'DELETE'
    })
    return this.request(request)
  }

  async patch<T>(
    url: URL | string,
    options: FetchModuleOptions = {}
  ): Promise<T> {
    const request = await this.createRequest(url, {
      ...options,
      method: 'PATCH'
    })
    return this.request<T>(request)
  }

  async put<T>(
    url: URL | string,
    options: FetchModuleOptions = {}
  ): Promise<T> {
    const request = await this.createRequest(url, { ...options, method: 'PUT' })
    return this.request<T>(request)
  }

  async post<T>(
    url: URL | string,
    options: FetchModuleOptions = {}
  ): Promise<T> {
    const request = await this.createRequest(url, {
      ...options,
      method: 'POST'
    })
    return this.request<T>(request)
  }

  /**
   * Convert an object to FormData, handling File objects and nested data
   * Uses lodash utilities for better type checking and iteration
   * @param data Object to convert to FormData
   * @returns FormData instance
   */
  private objectToFormData(data: Record<string, any>): FormData {
    const formData = new FormData()

    forOwn(data, (value, key) => {
      // Skip null/undefined values using lodash utilities
      if (isNull(value) || isUndefined(value)) {
        return
      }

      if (value instanceof File) {
        formData.append(key, value)
      } else if (isArray(value)) {
        forEach(value, (item, index) => {
          if (item instanceof File) {
            formData.append(`${key}[${index}]`, item)
          } else {
            formData.append(`${key}[${index}]`, toString(item))
          }
        })
      } else if (isObject(value)) {
        // Convert nested objects to JSON strings
        formData.append(key, JSON.stringify(value))
      } else {
        formData.append(key, toString(value))
      }
    })

    return formData
  }

  /**
   * POST method with automatic FormData conversion
   * Accepts any object and converts it to FormData, handling File objects properly
   * Automatically removes Content-Type header to let browser set multipart boundary
   * @param url URL to send the request to
   * @param data Object to convert to FormData
   * @param options Additional request options
   * @returns Promise resolving to the response data
   */
  async postFormData<T>(
    url: URL | string,
    data: Record<string, any>,
    options: FetchModuleOptions = {}
  ): Promise<T> {
    const formData = this.objectToFormData(data)

    const request = await this.createRequest(url, {
      ...options,
      method: 'POST',
      body: formData
    })

    return this.request<T>(request)
  }

  /**
   * PATCH method with automatic FormData conversion
   * Accepts any object and converts it to FormData, handling File objects properly
   * Automatically removes Content-Type header to let browser set multipart boundary
   * @param url URL to send the request to
   * @param data Object to convert to FormData
   * @param options Additional request options
   * @returns Promise resolving to the response data
   */
  async patchFormData<T>(
    url: URL | string,
    data: Record<string, any>,
    options: FetchModuleOptions = {}
  ): Promise<T> {
    const formData = this.objectToFormData(data)

    const request = await this.createRequest(url, {
      ...options,
      method: 'PATCH',
      body: formData
    })

    return this.request<T>(request)
  }
}
