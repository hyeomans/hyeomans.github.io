import assert from "node:assert/strict";
import test from "node:test";

test("login loads the CSRF token before a protected request", async () => {
  const handlers = new Map();
  const elements = new Map();

  for (const selector of [
    "#status",
    "#email",
    "#profile-name",
    "#register",
    "#login",
    "#load-session",
    "#save-profile",
    "#save-without-csrf",
    "#logout",
  ]) {
    elements.set(selector, {
      textContent: "",
      value: selector === "#email" ? "test@example.com" : "",
      addEventListener: (_event, handler) => handlers.set(selector, handler),
    });
  }

  globalThis.document = {
    querySelector: (selector) => elements.get(selector),
  };

  const bytes = new Uint8Array([1]).buffer;
  Object.defineProperty(globalThis, "navigator", {
    configurable: true,
    value: {
      credentials: {
        get: async () => ({
          id: "credential-id",
          rawId: bytes,
          type: "public-key",
          authenticatorAttachment: "platform",
          getClientExtensionResults: () => ({}),
          response: {
            clientDataJSON: bytes,
            authenticatorData: bytes,
            signature: bytes,
            userHandle: null,
          },
        }),
      },
    },
  });

  const expectedToken = "expected-csrf-token";
  let profileHeaders;
  globalThis.fetch = async (path, options = {}) => {
    if (path === "/api/passkeys/login/begin") {
      return jsonResponse({ publicKey: { challenge: "AQ", allowCredentials: [] } });
    }
    if (path === "/api/passkeys/login/finish") {
      return jsonResponse({ message: "signed in" });
    }
    if (path === "/api/session") {
      return jsonResponse({
        email: "test@example.com",
        profileName: "test@example.com",
        csrfToken: expectedToken,
      });
    }
    if (path === "/api/profile") {
      profileHeaders = options.headers;
      if (profileHeaders["X-CSRF-Token"] !== expectedToken) {
        return jsonResponse({ error: "missing or invalid CSRF token" }, 403);
      }
      return jsonResponse({ message: "profile updated" });
    }
    throw new Error(`unexpected request: ${path}`);
  };

  await import(`./app.js?test=${Date.now()}`);
  await handlers.get("#login")();
  await handlers.get("#save-profile")();

  assert.equal(profileHeaders["X-CSRF-Token"], expectedToken);
});

function jsonResponse(body, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
