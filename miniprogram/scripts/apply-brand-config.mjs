import fs from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const projectRoot = path.resolve(__dirname, '..')

async function readJson(filePath) {
  const content = await fs.readFile(filePath, 'utf8')
  return JSON.parse(content)
}

async function main() {
  const brandKey = (process.argv[2] || '').trim()
  if (!brandKey) {
    throw new Error('请传入品牌标识，例如：node scripts/apply-brand-config.mjs xunmeng')
  }

  const brandConfigPath = path.join(projectRoot, 'config', 'brands', `${brandKey}.json`)
  const manifestPath = path.join(projectRoot, 'src', 'manifest.json')
  const pagesPath = path.join(projectRoot, 'src', 'pages.json')

  const brandConfig = await readJson(brandConfigPath)
  const manifest = await readJson(manifestPath)
  const pages = await readJson(pagesPath)

  if (!brandConfig.appName || !brandConfig.appid || !brandConfig.envFile) {
    throw new Error(`品牌配置不完整：${brandConfigPath}`)
  }

  manifest.name = brandConfig.appName
  manifest.appid = brandConfig.appid
  manifest['mp-weixin'] = manifest['mp-weixin'] || {}
  manifest['mp-weixin'].appid = brandConfig.appid
  manifest.h5 = manifest.h5 || {}
  manifest.h5.title = brandConfig.appName

  pages.globalStyle = pages.globalStyle || {}
  pages.globalStyle.navigationBarTitleText = brandConfig.appName

  await fs.writeFile(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`, 'utf8')
  await fs.writeFile(pagesPath, `${JSON.stringify(pages, null, 2)}\n`, 'utf8')

  const envSourcePath = path.join(projectRoot, brandConfig.envFile)
  const envTargetPath = path.join(projectRoot, '.env.production')
  const envSourceContent = await fs.readFile(envSourcePath, 'utf8')
  const normalizedEnvContent = envSourceContent.replace(/\s+$/, '')
  const envContent = `${normalizedEnvContent}\nVITE_APP_NAME=${brandConfig.appName}\n`
  await fs.writeFile(envTargetPath, envContent, 'utf8')

  if (brandConfig.appid === 'wx0000000000000000') {
    console.warn(`[brand:${brandKey}] 当前使用占位 appid，请在发布前替换为真实 AppID`)
  }

  console.log(`[brand:${brandKey}] 已写入 manifest.json 和 .env.production`)
}

main().catch((error) => {
  console.error(error.message)
  process.exit(1)
})
