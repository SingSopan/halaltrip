import assert from "node:assert/strict";
import { afterEach, beforeEach, test } from "node:test";
import { APIError, apiRequest } from "../src/services/client.ts";

const originalFetch = globalThis.fetch;
const originalBaseURL = process.env.NEXT_PUBLIC_API_BASE_URL;

beforeEach(() => {
  process.env.NEXT_PUBLIC_API_BASE_URL = "http://localhost:8080/api/v1/";
});

afterEach(() => {
  globalThis.fetch = originalFetch;
  if (originalBaseURL === undefined) delete process.env.NEXT_PUBLIC_API_BASE_URL;
  else process.env.NEXT_PUBLIC_API_BASE_URL = originalBaseURL;
});

test("sends JSON, bearer token, and parses the backend JSON response", async () => {
  let request;
  globalThis.fetch = async (url, options) => {
    request = { url, options };
    return new Response(JSON.stringify({ user: { id: "user-1" } }), { status: 200, headers: { "Content-Type": "application/json" } });
  };
  const result = await apiRequest("/auth/login", { method: "POST", token: "test-token", body: { email: "a@example.com", password: "password" } });
  assert.equal(request.url, "http://localhost:8080/api/v1/auth/login");
  assert.equal(request.options.headers.get("Authorization"), "Bearer test-token");
  assert.equal(request.options.headers.get("Content-Type"), "application/json");
  assert.deepEqual(JSON.parse(request.options.body), { email: "a@example.com", password: "password" });
  assert.deepEqual(result, { user: { id: "user-1" } });
});

test("treats 204 responses as empty and preserves HTTP error details", async () => {
  globalThis.fetch = async () => new Response(null, { status: 204 });
  assert.equal(await apiRequest("/trips/one", { method: "DELETE" }), undefined);

  for (const status of [400, 401, 404, 500, 502, 504]) {
    globalThis.fetch = async () => new Response(JSON.stringify({ error: { code: `status_${status}`, message: "Pesan backend" } }), { status });
    await assert.rejects(apiRequest("/resource"), (error) => {
      assert.ok(error instanceof APIError);
      assert.equal(error.status, status);
      assert.equal(error.code, `status_${status}`);
      assert.equal(error.message, "Pesan backend");
      return true;
    });
  }
});

test("reports invalid JSON, network errors, and timeouts consistently", async () => {
  globalThis.fetch = async () => new Response("not json", { status: 200 });
  await assert.rejects(apiRequest("/resource"), (error) => error instanceof APIError && error.code === "invalid_response");

  globalThis.fetch = async () => { throw new TypeError("offline"); };
  await assert.rejects(apiRequest("/resource"), (error) => error instanceof APIError && error.code === "network_error");

  globalThis.fetch = async () => { throw new DOMException("deadline", "TimeoutError"); };
  await assert.rejects(apiRequest("/resource"), (error) => error instanceof APIError && error.code === "timeout");
});
