let csrfToken = "";

const statusBox = document.querySelector("#status");
const emailInput = document.querySelector("#email");
const profileNameInput = document.querySelector("#profile-name");

function show(value) {
  statusBox.textContent = typeof value === "string" ? value : JSON.stringify(value, null, 2);
}

function decodeBase64URL(value) {
  const base64 = value.replaceAll("-", "+").replaceAll("_", "/");
  const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
  return Uint8Array.from(atob(padded), (character) => character.charCodeAt(0));
}

function encodeBase64URL(value) {
  const bytes = new Uint8Array(value);
  let binary = "";
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary).replaceAll("+", "-").replaceAll("/", "_").replaceAll("=", "");
}

function creationOptionsFromJSON(options) {
  options.challenge = decodeBase64URL(options.challenge);
  options.user.id = decodeBase64URL(options.user.id);
  options.excludeCredentials = (options.excludeCredentials || []).map((credential) => ({
    ...credential,
    id: decodeBase64URL(credential.id),
  }));
  return options;
}

function requestOptionsFromJSON(options) {
  options.challenge = decodeBase64URL(options.challenge);
  options.allowCredentials = (options.allowCredentials || []).map((credential) => ({
    ...credential,
    id: decodeBase64URL(credential.id),
  }));
  return options;
}

function credentialToJSON(credential) {
  const response = {
    clientDataJSON: encodeBase64URL(credential.response.clientDataJSON),
  };
  if (credential.response.attestationObject) {
    response.attestationObject = encodeBase64URL(credential.response.attestationObject);
    response.transports = credential.response.getTransports?.() || [];
  } else {
    response.authenticatorData = encodeBase64URL(credential.response.authenticatorData);
    response.signature = encodeBase64URL(credential.response.signature);
    response.userHandle = credential.response.userHandle
      ? encodeBase64URL(credential.response.userHandle)
      : null;
  }
  return {
    id: credential.id,
    rawId: encodeBase64URL(credential.rawId),
    type: credential.type,
    authenticatorAttachment: credential.authenticatorAttachment,
    clientExtensionResults: credential.getClientExtensionResults(),
    response,
  };
}

async function api(path, options = {}) {
  const response = await fetch(path, options);
  const body = await response.json();
  if (!response.ok) throw new Error(`${response.status}: ${body.error}`);
  return body;
}

document.querySelector("#register").addEventListener("click", async () => {
  try {
    show("Starting registration…");
    const creation = await api("/api/passkeys/registration/begin", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: emailInput.value }),
    });
    const credential = await navigator.credentials.create({
      publicKey: creationOptionsFromJSON(creation.publicKey),
    });
    const result = await api("/api/passkeys/registration/finish", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(credentialToJSON(credential)),
    });
    show(result);
  } catch (error) {
    show(error.message);
  }
});

document.querySelector("#login").addEventListener("click", async () => {
  try {
    show("Starting sign-in…");
    const request = await api("/api/passkeys/login/begin", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: emailInput.value }),
    });
    const credential = await navigator.credentials.get({
      publicKey: requestOptionsFromJSON(request.publicKey),
    });
    const result = await api("/api/passkeys/login/finish", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(credentialToJSON(credential)),
    });
    const session = await api("/api/session");
    csrfToken = session.csrfToken;
    profileNameInput.value = session.profileName;
    show({
      ...result,
      session: { email: session.email, profileName: session.profileName },
    });
  } catch (error) {
    show(error.message);
  }
});

document.querySelector("#load-session").addEventListener("click", async () => {
  try {
    const session = await api("/api/session");
    csrfToken = session.csrfToken;
    profileNameInput.value = session.profileName;
    show({ email: session.email, profileName: session.profileName });
  } catch (error) {
    show(error.message);
  }
});

async function saveProfile(includeToken) {
  try {
    const headers = { "Content-Type": "application/json" };
    if (includeToken) headers["X-CSRF-Token"] = csrfToken;
    show(await api("/api/profile", {
      method: "POST",
      headers,
      body: JSON.stringify({ profileName: profileNameInput.value }),
    }));
  } catch (error) {
    show(error.message);
  }
}

document.querySelector("#save-profile").addEventListener("click", () => saveProfile(true));
document.querySelector("#save-without-csrf").addEventListener("click", () => saveProfile(false));
document.querySelector("#logout").addEventListener("click", async () => {
  try {
    show(await api("/api/logout", { method: "POST", headers: { "X-CSRF-Token": csrfToken } }));
    csrfToken = "";
  } catch (error) {
    show(error.message);
  }
});
