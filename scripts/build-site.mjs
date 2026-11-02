import { cp, mkdir, readFile, readdir, rename, rm, writeFile } from "node:fs/promises";
import path from "node:path";

const root = process.cwd();
const docs = path.join(root, ".site-docs");
const generatedConfig = path.join(root, ".site-mkdocs.yml");
const topics = [
  "Go学习笔记",
  "区块链学习笔记_38篇",
  "智能合约学习笔记",
  "钱包学习笔记",
];

if (path.dirname(docs) !== root || path.basename(docs) !== ".site-docs") {
  throw new Error("站点输出目录不在仓库根目录内");
}
await rm(docs, { recursive: true, force: true });
await mkdir(docs, { recursive: true });

function siteLinks(markdown) {
  return markdown.replace(/\]\(([^)]+)\)/g, (whole, target) => {
    if (/^(?:https?:|mailto:|#)/i.test(target)) return whole;
    const [file, anchor] = target.split("#", 2);
    if (!file.endsWith("README.md")) return whole;
    return "](" + file.slice(0, -"README.md".length) + "index.md" +
      (anchor ? "#" + anchor : "") + ")";
  });
}

async function transformMarkdown(directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const full = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      await transformMarkdown(full);
    } else if (entry.isFile() && entry.name.endsWith(".md")) {
      const before = await readFile(full, "utf8");
      const after = siteLinks(before);
      if (before !== after) await writeFile(full, after, "utf8");
    }
  }
}

await writeFile(path.join(docs, "index.md"), siteLinks(await readFile(path.join(root, "README.md"), "utf8")));
await cp(path.join(root, "CONTRIBUTING.md"), path.join(docs, "CONTRIBUTING.md"));
for (const topic of topics) {
  const target = path.join(docs, topic);
  await cp(path.join(root, topic), target, { recursive: true });
  await rename(path.join(target, "README.md"), path.join(target, "index.md"));
}
await transformMarkdown(docs);

const repo = "https://github.com/FortresseAssiegee/web3Note/tree/main/";
const codePages = [
  ["Go学习笔记/examples", "Go 独立示例"],
  ["Go学习笔记/code", "Go 配套工程"],
  ["区块链学习笔记_38篇/附加资料/代码示例", "区块链配套代码"],
];
for (const [folder, title] of codePages) {
  const text = "# " + title + "\n\n" +
    "代码与运行说明保存在主仓库：[在 GitHub 浏览源文件](" +
    repo + encodeURI(folder) + ")。\n";
  await writeFile(path.join(docs, folder, "index.md"), text, "utf8");
}

const mathjax = [
  "window.MathJax = {",
  "  tex: { inlineMath: [['\\\\(', '\\\\)']], displayMath: [['\\\\[', '\\\\]']], processEscapes: true },",
  "  options: { ignoreHtmlClass: '.*|', processHtmlClass: 'arithmatex' }",
  "};",
  "document$.subscribe(() => {",
  "  MathJax.startup.output.clearCache();",
  "  MathJax.typesetClear();",
  "  MathJax.texReset();",
  "  MathJax.typesetPromise();",
  "});",
  "",
].join("\n");
await mkdir(path.join(docs, "javascripts"), { recursive: true });
await writeFile(path.join(docs, "javascripts", "mathjax.js"), mathjax, "utf8");

function title(file) {
  return file.replace(/\.md$/, "").replace(/^\d{2}_/, "").replaceAll("_", " ");
}
function item(label, value, depth = 1) {
  const indent = "  ".repeat(depth);
  if (Array.isArray(value)) {
    return indent + "- " + JSON.stringify(label) + ":\n" +
      value.map(([childLabel, childValue]) => item(childLabel, childValue, depth + 1)).join("");
  }
  return indent + "- " + JSON.stringify(label) + ": " + JSON.stringify(value) + "\n";
}
async function numberedPages(folder) {
  return (await readdir(path.join(docs, folder)))
    .filter((name) => /^\d{2}_.*\.md$/.test(name))
    .sort((a, b) => a.localeCompare(b, "zh", { numeric: true }));
}
const goPages = await numberedPages("Go学习笔记");
const goNav = [
  ["概览", "Go学习笔记/index.md"],
  ...goPages.map((name) => [title(name), "Go学习笔记/" + name]),
  ["代码示例", "Go学习笔记/examples/index.md"],
  ["配套工程", "Go学习笔记/code/index.md"],
];
const blockchain = "区块链学习笔记_38篇";
const partNames = (await readdir(path.join(docs, blockchain), { withFileTypes: true }))
  .filter((entry) => entry.isDirectory() && entry.name.startsWith("第"))
  .map((entry) => entry.name);
const parts = await Promise.all(partNames.map(async (name) => ({
  name,
  files: await numberedPages(path.join(blockchain, name)),
})));
parts.sort((a, b) => Number(a.files[0]?.slice(0, 2)) - Number(b.files[0]?.slice(0, 2)));
const chainNav = [
  ["概览", blockchain + "/index.md"],
  ...parts.map(({ name, files }) => [name, files.map((file) => [title(file), blockchain + "/" + name + "/" + file])]),
  ["代码示例", blockchain + "/附加资料/代码示例/index.md"],
];
const nav = [
  ["首页", "index.md"],
  ["Go 学习笔记", goNav],
  ["区块链与 Web3", chainNav],
  ["智能合约（待整理）", "智能合约学习笔记/index.md"],
  ["钱包（待整理）", "钱包学习笔记/index.md"],
  ["贡献与维护", "CONTRIBUTING.md"],
];
const config = (await readFile(path.join(root, "mkdocs.yml"), "utf8")).trimEnd() +
  "\nnav:\n" + nav.map(([label, value]) => item(label, value)).join("");
await writeFile(generatedConfig, config, "utf8");
console.log("站点源文件已生成：Go " + goPages.filter((name) => !name.startsWith("00_")).length + " 篇，区块链 " +
  parts.reduce((sum, part) => sum + part.files.length, 0) + " 篇。");
