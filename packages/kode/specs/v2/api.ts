// @ts-nocheck

import { Kode } from "@kode/core"
import { ReadTool } from "@kode/core/tools"

const kode = Kode.make({})

kode.tool.add(ReadTool)

kode.tool.add({
  name: "bash",
  schema: {
    type: "object",
    properties: {
      command: {
        type: "string",
        description: "The command to run.",
      },
    },
    required: ["command"],
  },
  execute(input, ctx) {},
})

kode.auth.add({
  provider: "openai",
  type: "api",
  value: process.env.OPENAI_API_KEY,
})

kode.agent.add({
  name: "build",
  permissions: [],
  model: {
    id: "gpt-5-5",
    provider: "openai",
    variant: "xhigh",
  },
})

const sessionID = await kode.session.create({
  agent: "build",
})

kode.subscribe((event) => {
  console.log(event)
})

await kode.session.prompt({
  sessionID,
  text: "hey what is up",
})

await kode.session.prompt({
  sessionID,
  text: "what is up with this",
  files: [
    {
      mime: "image/png",
      uri: "data:image/png;base64,xxxx",
    },
  ],
})

await kode.session.wait()

console.log(await kode.session.messages(sessionID))
