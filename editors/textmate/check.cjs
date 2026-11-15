// Tokenizes every `narra` fence of the given markdown files (default: the guide and
// SPEC.md) and every given .nra file. Fails on an `invalid` scope, or on a string,
// template or comment still open at the end of a sample.
const fs = require("fs")
const path = require("path")
const textmate = require("vscode-textmate")
const oniguruma = require("vscode-oniguruma")

const root = path.resolve(__dirname, "..", "..")
const grammarPath = path.join(__dirname, "narra.tmLanguage.json")
const wasmPath = require.resolve("vscode-oniguruma/release/onig.wasm")

function collectSamples(file) {
  const text = fs.readFileSync(file, "utf8")
  if (file.endsWith(".nra")) {
    return [{ file, line: 1, code: text }]
  }
  const samples = []
  const lines = text.split("\n")
  let current = null
  for (let index = 0; index < lines.length; index++) {
    const line = lines[index]
    if (current === null) {
      const open = /^```(\S*)/.exec(line)
      if (open !== null && open[1] === "narra") {
        current = { line: index + 2, code: [], skip: false }
      } else if (open !== null) {
        current = { skip: true }
      }
    } else if (line.startsWith("```")) {
      if (!current.skip) {
        samples.push({ file, line: current.line, code: current.code.join("\n") })
      }
      current = null
    } else if (!current.skip) {
      current.code.push(line)
    }
  }
  return samples
}

function listFiles(requested) {
  if (requested.length > 0) {
    return requested.map((item) => path.resolve(item))
  }
  const guide = path.join(root, "docs", "guide")
  const files = fs
    .readdirSync(guide)
    .filter((name) => name.endsWith(".md"))
    .map((name) => path.join(guide, name))
  files.push(path.join(root, "SPEC.md"))
  return files
}

async function main() {
  await oniguruma.loadWASM(fs.readFileSync(wasmPath).buffer)
  const registry = new textmate.Registry({
    onigLib: Promise.resolve({
      createOnigScanner: (patterns) => new oniguruma.OnigScanner(patterns),
      createOnigString: (text) => new oniguruma.OnigString(text),
    }),
    loadGrammar: async (scopeName) => {
      if (scopeName !== "source.narra") {
        return null
      }
      return textmate.parseRawGrammar(fs.readFileSync(grammarPath, "utf8"), grammarPath)
    },
  })
  const grammar = await registry.loadGrammar("source.narra")
  let sampleCount = 0
  const problems = []
  for (const file of listFiles(process.argv.slice(2))) {
    const shortName = path.relative(root, file)
    for (const sample of collectSamples(file)) {
      sampleCount++
      let stack = textmate.INITIAL
      const codeLines = sample.code.split("\n")
      for (let offset = 0; offset < codeLines.length; offset++) {
        const result = grammar.tokenizeLine(codeLines[offset], stack)
        for (const token of result.tokens) {
          if (token.scopes.some((scope) => scope.startsWith("invalid"))) {
            const text = codeLines[offset].slice(token.startIndex, token.endIndex)
            problems.push(`${shortName}:${sample.line + offset}: invalid scope on "${text}"`)
          }
        }
        stack = result.ruleStack
      }
      if (stack.depth > 1) {
        problems.push(`${shortName}:${sample.line}: the sample ends inside an open string, template or comment`)
      }
    }
  }
  for (const problem of problems) {
    console.log(problem)
  }
  console.log(`${sampleCount} samples, ${problems.length} problems`)
  process.exit(problems.length === 0 ? 0 : 1)
}

main()
