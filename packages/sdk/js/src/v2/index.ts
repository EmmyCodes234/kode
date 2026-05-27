export * from "./client.js"
export * from "./server.js"

import { createKodeClient } from "./client.js"
import { createKodeServer } from "./server.js"
import type { ServerOptions } from "./server.js"

export * as data from "./data.js"

export async function createkode(options?: ServerOptions) {
  const server = await createKodeServer({
    ...options,
  })

  const client = createKodeClient({
    baseUrl: server.url,
  })

  return {
    client,
    server,
  }
}
