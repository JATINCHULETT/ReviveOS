import * as http from "http";
import * as https from "https";
import { URL } from "url";
import { ReviveOSAuthError, ReviveOSApiError, ReviveOSValidationError } from "./errors";

export interface HttpClientOptions {
  apiKey?: string;
  endpoint: string;
  timeoutMs?: number;
  maxRetries?: number;
}

export class ReviveOSHttpClient {
  private readonly endpoint: string;
  private readonly apiKey?: string;
  private readonly timeoutMs: number;
  private readonly maxRetries: number;

  constructor(options: HttpClientOptions) {
    this.endpoint = options.endpoint.replace(/\/$/, "");
    this.apiKey = options.apiKey;
    this.timeoutMs = options.timeoutMs || 10000;
    this.maxRetries = options.maxRetries ?? 2;
  }

  /**
   * Execute an HTTP request with automatic retries, backoff, and JSON parsing.
   */
  public async request<T = any>(
    path: string,
    method: "GET" | "POST" | "PUT" | "DELETE" = "GET",
    body?: any,
    headers: Record<string, string> = {}
  ): Promise<T> {
    const urlString = `${this.endpoint}${path.startsWith("/") ? path : `/${path}`}`;
    const targetUrl = new URL(urlString);

    const defaultHeaders: Record<string, string> = {
      "Content-Type": "application/json",
      "Accept": "application/json",
      "User-Agent": "@reviveos/razorpay-sdk/1.0.0",
      ...headers,
    };

    if (this.apiKey) {
      defaultHeaders["Authorization"] = `Bearer ${this.apiKey}`;
      defaultHeaders["X-API-Key"] = this.apiKey;
    }

    let lastError: Error | null = null;
    const bodyData = body ? JSON.stringify(body) : undefined;
    if (bodyData) {
      defaultHeaders["Content-Length"] = Buffer.byteLength(bodyData).toString();
    }

    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        if (attempt > 0) {
          // Exponential backoff with jitter: 200ms, 400ms, 800ms
          const backoff = Math.min(200 * Math.pow(2, attempt - 1) + Math.random() * 50, 3000);
          await new Promise((resolve) => setTimeout(resolve, backoff));
        }

        const response = await this.executeHttpRequest(targetUrl, method, defaultHeaders, bodyData);

        if (response.statusCode >= 200 && response.statusCode < 300) {
          if (!response.body || response.body.trim() === "") {
            return {} as T;
          }
          try {
            return JSON.parse(response.body) as T;
          } catch {
            return response.body as unknown as T;
          }
        }

        if (response.statusCode === 401 || response.statusCode === 403) {
          throw new ReviveOSAuthError(`ReviveOS authentication failed (${response.statusCode}): ${response.body}`);
        }

        if (response.statusCode === 400) {
          throw new ReviveOSValidationError(`ReviveOS bad request: ${response.body}`);
        }

        // 5xx Server errors are eligible for retry
        if (response.statusCode >= 500 && attempt < this.maxRetries) {
          lastError = new ReviveOSApiError(response.body, response.statusCode);
          continue;
        }

        throw new ReviveOSApiError(response.body, response.statusCode);
      } catch (err: any) {
        lastError = err;
        if (err instanceof ReviveOSAuthError || err instanceof ReviveOSValidationError) {
          throw err;
        }
        if (attempt >= this.maxRetries) {
          throw err;
        }
      }
    }

    throw lastError || new Error("Request to ReviveOS failed after maximum retries");
  }

  private executeHttpRequest(
    url: URL,
    method: string,
    headers: Record<string, string>,
    body?: string
  ): Promise<{ statusCode: number; body: string }> {
    return new Promise((resolve, reject) => {
      const isHttps = url.protocol === "https:";
      const client = isHttps ? https : http;

      const req = client.request(
        url,
        {
          method,
          headers,
          timeout: this.timeoutMs,
        },
        (res) => {
          let data = "";
          res.on("data", (chunk) => {
            data += chunk;
          });
          res.on("end", () => {
            resolve({
              statusCode: res.statusCode || 500,
              body: data,
            });
          });
        }
      );

      req.on("timeout", () => {
        req.destroy();
        reject(new Error(`ReviveOS request timed out after ${this.timeoutMs}ms`));
      });

      req.on("error", (err) => {
        reject(err);
      });

      if (body) {
        req.write(body);
      }
      req.end();
    });
  }
}
