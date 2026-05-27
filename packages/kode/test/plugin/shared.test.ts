import { describe, expect, test } from "bun:test"
import { parsePluginSpecifier } from "../../src/plugin/shared"

describe("parsePluginSpecifier", () => {
  test("parses standard npm package without version", () => {
    expect(parsePluginSpecifier("acme")).toEqual({
      pkg: "acme",
      version: "latest",
    })
  })

  test("parses standard npm package with version", () => {
    expect(parsePluginSpecifier("acme@1.0.0")).toEqual({
      pkg: "acme",
      version: "1.0.0",
    })
  })

  test("parses scoped npm package without version", () => {
    expect(parsePluginSpecifier("@kode/acme")).toEqual({
      pkg: "@kode/acme",
      version: "latest",
    })
  })

  test("parses scoped npm package with version", () => {
    expect(parsePluginSpecifier("@kode/acme@1.0.0")).toEqual({
      pkg: "@kode/acme",
      version: "1.0.0",
    })
  })

  test("parses package with git+https url", () => {
    expect(parsePluginSpecifier("acme@git+https://github.com/kode/acme.git")).toEqual({
      pkg: "acme",
      version: "git+https://github.com/kode/acme.git",
    })
  })

  test("parses scoped package with git+https url", () => {
    expect(parsePluginSpecifier("@kode/acme@git+https://github.com/kode/acme.git")).toEqual({
      pkg: "@kode/acme",
      version: "git+https://github.com/kode/acme.git",
    })
  })

  test("parses package with git+ssh url containing another @", () => {
    expect(parsePluginSpecifier("acme@git+ssh://git@github.com/kode/acme.git")).toEqual({
      pkg: "acme",
      version: "git+ssh://git@github.com/kode/acme.git",
    })
  })

  test("parses scoped package with git+ssh url containing another @", () => {
    expect(parsePluginSpecifier("@kode/acme@git+ssh://git@github.com/kode/acme.git")).toEqual({
      pkg: "@kode/acme",
      version: "git+ssh://git@github.com/kode/acme.git",
    })
  })

  test("parses unaliased git+ssh url", () => {
    expect(parsePluginSpecifier("git+ssh://git@github.com/kode/acme.git")).toEqual({
      pkg: "git+ssh://git@github.com/kode/acme.git",
      version: "",
    })
  })

  test("parses npm alias using the alias name", () => {
    expect(parsePluginSpecifier("acme@npm:@kode/acme@1.0.0")).toEqual({
      pkg: "acme",
      version: "npm:@kode/acme@1.0.0",
    })
  })

  test("parses bare npm protocol specifier using the target package", () => {
    expect(parsePluginSpecifier("npm:@kode/acme@1.0.0")).toEqual({
      pkg: "@kode/acme",
      version: "1.0.0",
    })
  })

  test("parses unversioned npm protocol specifier", () => {
    expect(parsePluginSpecifier("npm:@kode/acme")).toEqual({
      pkg: "@kode/acme",
      version: "latest",
    })
  })
})
