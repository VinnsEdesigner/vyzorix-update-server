import axios, { type AxiosError } from 'axios';

export const ApiErrorCode = {
  BAD_REQUEST: 'bad_request',
  UNAUTHORIZED: 'unauthorized',
  FORBIDDEN: 'forbidden',
  NOT_FOUND: 'not_found',
  CONFLICT: 'conflict',
  RATE_LIMIT: 'rate_limit_exceeded',
  INTERNAL_ERROR: 'internal_error',
  INVALID_INPUT: 'invalid_input',
} as const;

export type ApiErrorCode = typeof ApiErrorCode[keyof typeof ApiErrorCode];

export interface RateLimitInfo {
  limit: number;
  remaining: number;
  resetAt: number;
  retryAfter: number;
}

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly code: ApiErrorCode,
    public readonly statusCode: number,
    public readonly details?: Record<string, unknown>,
    public readonly rateLimit?: RateLimitInfo
  ) {
    super(message);
    this.name = 'ApiError';
  }

  isUnauthorized(): boolean {
    return this.statusCode === 401;
  }

  isForbidden(): boolean {
    return this.statusCode === 403;
  }

  isNotFound(): boolean {
    return this.statusCode === 404;
  }

  isRateLimited(): boolean {
    return this.statusCode === 429;
  }

  isServerError(): boolean {
    return this.statusCode >= 500;
  }
}

function parseRateLimitHeaders(headers: Record<string, string>): RateLimitInfo | undefined {
  const limit = headers['x-ratelimit-limit'];
  const remaining = headers['x-ratelimit-remaining'];
  const reset = headers['x-ratelimit-reset'];
  const retryAfter = headers['retry-after'];

  if (limit || remaining || reset) {
    return {
      limit: limit ? parseInt(limit, 10) : 0,
      remaining: remaining ? parseInt(remaining, 10) : 0,
      resetAt: reset ? parseInt(reset, 10) : 0,
      retryAfter: retryAfter ? parseInt(retryAfter, 10) : 0,
    };
  }
  return undefined;
}

export function parseApiError(error: unknown): ApiError {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{ error?: string; message?: string }>;
    const statusCode = axiosError.response?.status ?? 0;
    const responseData = axiosError.response?.data;
    const headers = axiosError.response?.headers as Record<string, string> | undefined;

        let code: ApiErrorCode = ApiErrorCode.INTERNAL_ERROR;
    let message = 'An unexpected error occurred';
    let details: Record<string, unknown> | undefined;
    let rateLimit: RateLimitInfo | undefined;

    if (responseData) {
      code = (responseData.error as ApiErrorCode) || getCodeFromStatus(statusCode);
      message = responseData.message || getMessageFromCode(code);
      details = responseData;
    }

        if (headers) {
      rateLimit = parseRateLimitHeaders(headers);
    }

        if (details?.error === 'account_locked') {
      return new ApiError(
        (details.message as string) || 'Account temporarily locked',
        ApiErrorCode.FORBIDDEN,
        403,
        {
          ...details,
          locked_until: details.locked_until,
          retry_after: details.retry_after,
        }
      );
    }

    return new ApiError(message, code, statusCode, details, rateLimit);
  }

    if (error instanceof Error) {
    return new ApiError(error.message, ApiErrorCode.INTERNAL_ERROR, 500);
  }

  return new ApiError('An unexpected error occurred', ApiErrorCode.INTERNAL_ERROR, 500);
}

function getCodeFromStatus(status: number): ApiErrorCode {
  switch (status) {
    case 400: return ApiErrorCode.BAD_REQUEST;
    case 401: return ApiErrorCode.UNAUTHORIZED;
    case 403: return ApiErrorCode.FORBIDDEN;
    case 404: return ApiErrorCode.NOT_FOUND;
    case 409: return ApiErrorCode.CONFLICT;
    case 429: return ApiErrorCode.RATE_LIMIT;
    case 500:
    case 502:
    case 503:
    default: return ApiErrorCode.INTERNAL_ERROR;
  }
}

function getMessageFromCode(code: string): string {
  const messages: Record<string, string> = {
    [ApiErrorCode.BAD_REQUEST]: 'The request was invalid',
    [ApiErrorCode.UNAUTHORIZED]: 'Authentication required',
    [ApiErrorCode.FORBIDDEN]: 'Access denied',
    [ApiErrorCode.NOT_FOUND]: 'Resource not found',
    [ApiErrorCode.CONFLICT]: 'Resource already exists',
    [ApiErrorCode.INTERNAL_ERROR]: 'An unexpected error occurred',
    [ApiErrorCode.RATE_LIMIT]: 'Too many requests, please try again later',
    [ApiErrorCode.INVALID_INPUT]: 'Invalid input provided',
  };
  return messages[code] || 'An error occurred';
}

export async function withErrorHandling<T>(
  apiCall: () => Promise<T>
): Promise<T> {
  try {
    return await apiCall();
  } catch (error) {
    throw parseApiError(error);
  }
}
