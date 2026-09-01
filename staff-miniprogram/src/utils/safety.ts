/**
 * 服务过程安全监控（阶段三）
 * 签到成功后启动：录音 + 60s 定位上报；签退时停止并上传录音。
 * 模块级单例，跨页面切换保持运行。
 */
import { staffSafetyApi, uploadServiceAudio } from '@/api'

const recorderManager = uni.getRecorderManager()

let activeOrderId: string | number | null = null
let locationTimer: ReturnType<typeof setInterval> | null = null
let audioFilePath = ''
let manuallyStopped = false

// 录音停止回调：手动停止时保留文件；自动到时则重启续录（保留最新片段）
recorderManager.onStop((res: any) => {
  audioFilePath = res?.tempFilePath || ''
  if (!manuallyStopped && activeOrderId !== null) {
    // 达到最大时长自动停止：立即重启续录
    try {
      recorderManager.start({ format: 'mp3', duration: 600000 })
    } catch {
      // 录音重启失败不影响定位上报
    }
  }
})

/** 当前是否处于服务安全监控中 */
export function isSafetyActive(): boolean {
  return activeOrderId !== null
}

/** 上报一次定位（失败静默，下次重试） */
async function reportLocation(orderId: string | number) {
  try {
    const loc: any = await new Promise((resolve, reject) => {
      uni.getLocation({ type: 'gcj02', success: resolve, fail: reject })
    })
    await staffSafetyApi.reportLocation(orderId, loc.latitude, loc.longitude)
  } catch {
    // 静默失败
  }
}

/**
 * 启动服务安全监控（签到成功后调用）
 * - 启动录音（服务过程留痕）
 * - 每 60s 上报一次定位
 */
export function startSafety(orderId: string | number) {
  if (activeOrderId === orderId) return
  activeOrderId = orderId
  manuallyStopped = false
  audioFilePath = ''

  // 启动录音
  try {
    recorderManager.start({ format: 'mp3', duration: 600000 })
  } catch {
    // 录音失败不阻塞服务
  }

  // 定位上报：立即一次 + 每 60s 一次
  reportLocation(orderId)
  locationTimer = setInterval(() => {
    if (activeOrderId === orderId) {
      reportLocation(orderId)
    }
  }, 60 * 1000)
}

/**
 * 停止服务安全监控（签退时调用）
 * 停止录音并上传最新片段，返回已提交的录音 URL（失败返回 null，不阻塞签退）
 */
export async function stopSafety(): Promise<string | null> {
  const orderId = activeOrderId
  activeOrderId = null
  if (locationTimer) {
    clearInterval(locationTimer)
    locationTimer = null
  }
  manuallyStopped = true
  if (orderId === null) return null

  // 停止录音并等待文件就绪
  try {
    recorderManager.stop()
  } catch {
    // 忽略
  }
  await new Promise((resolve) => setTimeout(resolve, 600))

  if (!audioFilePath) return null
  try {
    const { url } = await uploadServiceAudio(audioFilePath)
    await staffSafetyApi.submitAudio(orderId, url)
    audioFilePath = ''
    return url
  } catch {
    // 录音上传失败不阻塞签退主流程
    return null
  }
}
