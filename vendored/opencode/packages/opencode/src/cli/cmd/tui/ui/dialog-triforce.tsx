import { TextAttributes } from "@opentui/core"
import { useTheme } from "../context/theme"
import { useDialog, type DialogContext } from "./dialog"
import { createStore } from "solid-js/store"
import { For } from "solid-js"
import { useBindings } from "../keymap"

export type DialogTriforceResult = "retry" | "force" | "abort" | undefined

export type DialogTriforceProps = {
  title: string
  message: string
  details: string
  onRetry?: () => void
  onForce?: () => void
  onAbort?: () => void
}

const OPTIONS = ["retry", "force", "abort"] as const

export function DialogTriforce(props: DialogTriforceProps) {
  const dialog = useDialog()
  const { theme } = useTheme()
  const [store, setStore] = createStore({
    active: 0,
  })

  useBindings(() => ({
    bindings: [
      {
        key: "return",
        desc: "Confirm dialog selection",
        group: "Dialog",
        cmd: () => {
          const choice = OPTIONS[store.active]
          if (choice === "retry") props.onRetry?.()
          else if (choice === "force") props.onForce?.()
          else props.onAbort?.()
          dialog.clear()
        },
      },
      {
        key: "left",
        desc: "Previous dialog option",
        group: "Dialog",
        cmd: () => {
          setStore("active", store.active === 0 ? OPTIONS.length - 1 : store.active - 1)
        },
      },
      {
        key: "right",
        desc: "Next dialog option",
        group: "Dialog",
        cmd: () => {
          setStore("active", (store.active + 1) % OPTIONS.length)
        },
      },
    ],
  }))

  return (
    <box paddingLeft={2} paddingRight={2} gap={1}>
      <box flexDirection="row" justifyContent="space-between">
        <text attributes={TextAttributes.BOLD} fg={theme.text}>
          {props.title}
        </text>
        <text fg={theme.textMuted} onMouseUp={() => dialog.clear()}>
          esc
        </text>
      </box>
      <box paddingBottom={1}>
        <text fg={theme.textMuted}>{props.message}</text>
      </box>
      <box paddingBottom={1}>
        <text fg={theme.textMuted}>{props.details}</text>
      </box>
      <box flexDirection="row" justifyContent="flex-end" paddingBottom={1}>
        <For each={OPTIONS}>
          {(key, index) => (
            <box
              paddingLeft={1}
              paddingRight={1}
              backgroundColor={index() === store.active ? theme.primary : undefined}
              onMouseUp={() => {
                if (key === "retry") props.onRetry?.()
                else if (key === "force") props.onForce?.()
                else props.onAbort?.()
                dialog.clear()
              }}
            >
              <text fg={index() === store.active ? theme.selectedListItemText : theme.textMuted}>
                {key}
              </text>
            </box>
          )}
        </For>
      </box>
    </box>
  )
}

DialogTriforce.show = (dialog: DialogContext, title: string, message: string, details: string) => {
  return new Promise<DialogTriforceResult>((resolve) => {
    dialog.replace(
      () => (
        <DialogTriforce
          title={title}
          message={message}
          details={details}
          onRetry={() => resolve("retry")}
          onForce={() => resolve("force")}
          onAbort={() => resolve("abort")}
        />
      ),
      () => resolve(undefined),
    )
  })
}
