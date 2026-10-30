// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

// TEMPORARY demo shim, not for the PR. It stands in for the SDK change: it collects the `sessionData`
// every /flow/execute response returns (INCOMPLETE included) and adds it to the SDK's
// /oauth2/auth/callback POST, so a code requested with `device_sso` can return a device secret.

interface HeldSession {
  flowIdentifier: string;
  sessionHandle: string;
}

const held: Map<string, HeldSession> = new Map();

const requestUrl = (input: RequestInfo | URL): string => {
  if (typeof input === 'string') return input;
  if (input instanceof URL) return input.href;
  return input.url;
};

export default function installSessionDataCallbackShim(): void {
  if (typeof window === 'undefined') return;
  const originalFetch: typeof window.fetch = window.fetch.bind(window);

  window.fetch = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
    const url: string = requestUrl(input);
    let requestInit: RequestInit | undefined = init;

    if (url.endsWith('/oauth2/auth/callback') && typeof init?.body === 'string' && held.size > 0) {
      const body = JSON.parse(init.body) as Record<string, unknown>;
      body['sessionData'] = [...held.values()];
      requestInit = {...init, body: JSON.stringify(body)};
    }

    const response: Response = await originalFetch(input, requestInit);

    if (url.includes('/flow/execute')) {
      // Read before handing the response back, so a session is held before the SDK posts the callback.
      const data = (await response
        .clone()
        .json()
        .catch(() => null)) as {sessionData?: HeldSession[]; clearedSessions?: string[]} | null;
      data?.sessionData?.forEach((s: HeldSession) => {
        held.set(s.flowIdentifier, {flowIdentifier: s.flowIdentifier, sessionHandle: s.sessionHandle});
      });
      data?.clearedSessions?.forEach((flowIdentifier: string) => held.delete(flowIdentifier));
    }

    return response;
  };
}
