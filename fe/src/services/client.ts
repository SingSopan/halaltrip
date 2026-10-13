export type APIErrorPayload = { error?: { code?: string; message?: string } };

export class APIError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(message: string, status: number, code: string) {
    super(message);
    this.name = "APIError";
    this.status = status;
    this.code = code;
  }
}

function getBaseURL() {
  const value = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (!value) {
    throw new APIError(
      "NEXT_PUBLIC_API_BASE_URL belum dikonfigurasi. Isi URL backend di fe/.env.local lalu restart Next.js.",
      0,
      "api_not_configured",
    );
  }
  return value.replace(/\/+$/, "");
}

export async function apiRequest<T>(
  path: string,
  options: {
    method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
    token?: string | null;
    body?: unknown;
    signal?: AbortSignal;
  } = {},
): Promise<T> {
  const headers = new Headers({ Accept: "application/json" });
  if (options.body !== undefined) headers.set("Content-Type", "application/json");
  if (options.token) headers.set("Authorization", `Bearer ${options.token}`);

  let response: Response;
  try {
    response = await fetch(`${getBaseURL()}${path}`, {
      method: options.method ?? "GET",
      headers,
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
      signal: options.signal ?? AbortSignal.timeout(15_000),
      cache: "no-store",
    });
  } catch (error) {
    if (error instanceof APIError) throw error;
    if (error instanceof DOMException && error.name === "TimeoutError") {
      throw new APIError("Permintaan ke backend melewati batas waktu. Coba lagi.", 0, "timeout");
    }
    throw new APIError("Backend tidak dapat dijangkau. Pastikan backend hidup dan URL API benar.", 0, "network_error");
  }

  if (response.status === 204) return undefined as T;
  const text = await response.text();
  let payload: unknown;
  if (text) {
    try {
      payload = JSON.parse(text) as unknown;
    } catch {
      if (response.ok) throw new APIError("Backend mengirim respons yang bukan JSON valid.", response.status, "invalid_response");
    }
  }
  if (!response.ok) {
    if (response.status === 401 && options.token) {
      if (typeof window !== "undefined") window.dispatchEvent(new Event("halaltrip:auth-expired"));
    }
    const error = payload && typeof payload === "object" ? (payload as APIErrorPayload).error : undefined;
    throw new APIError(error?.message ?? `Permintaan gagal (HTTP ${response.status}).`, response.status, error?.code ?? "http_error");
  }
  return payload as T;
}
