const defaultBaseUrl = "http://localhost:8080";

export async function apiRequest(path, options = {}) {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || defaultBaseUrl;
  const { method = "GET", body, token } = options;
  const headers = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${baseUrl}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const message = await res.text();
    throw new Error(message || "Request failed");
  }

  if (res.status === 204) {
    return null;
  }
  return res.json();
}
