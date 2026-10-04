// 100% Offline Web Audio API Synthesizer (Zero external audio file downloads)

let audioCtx: AudioContext | null = null

const SOUND_ENABLED_KEY = 'local_finance_sound_enabled'
const SOUND_VOLUME_KEY = 'local_finance_sound_volume'

export function isSoundEnabled(): boolean {
  if (typeof window === 'undefined') return true
  const stored = localStorage.getItem(SOUND_ENABLED_KEY)
  return stored === null ? true : stored === 'true'
}

export function setSoundEnabled(enabled: boolean): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(SOUND_ENABLED_KEY, String(enabled))
}

export function getSoundVolume(): number {
  if (typeof window === 'undefined') return 0.3
  const stored = localStorage.getItem(SOUND_VOLUME_KEY)
  if (!stored) return 0.3
  const val = parseFloat(stored)
  return isNaN(val) ? 0.3 : Math.max(0.05, Math.min(val, 1.0))
}

export function setSoundVolume(vol: number): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(SOUND_VOLUME_KEY, String(Math.max(0.05, Math.min(vol, 1.0))))
}

function getAudioContext(): AudioContext | null {
  if (typeof window === 'undefined') return null
  if (!audioCtx) {
    const AudioContextClass =
      window.AudioContext ||
      (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    if (AudioContextClass) {
      audioCtx = new AudioContextClass()
    }
  }
  if (audioCtx && audioCtx.state === 'suspended') {
    audioCtx.resume().catch(() => {})
  }
  return audioCtx
}

// Automatically unlock audio context on first click or touch
if (typeof window !== 'undefined') {
  const unlockAudio = () => {
    const ctx = getAudioContext()
    if (ctx && ctx.state === 'suspended') {
      ctx.resume().catch(() => {})
    }
  }
  window.addEventListener('pointerdown', unlockAudio, { capture: true, passive: true })
  window.addEventListener('keydown', unlockAudio, { capture: true, passive: true })
}

/**
 * Play a gentle, clear harmonic chime for successful operations and milestones
 */
export function playSuccessChime() {
  if (!isSoundEnabled()) return

  try {
    const ctx = getAudioContext()
    if (!ctx) return

    if (ctx.state === 'suspended') {
      ctx.resume()
    }

    const volume = getSoundVolume()
    const now = ctx.currentTime

    // 3-note harmonic arpeggio: C5 (523Hz) -> E5 (659Hz) -> G5 (784Hz)
    const notes = [
      { freq: 523.25, time: now, dur: 0.2 },
      { freq: 659.25, time: now + 0.08, dur: 0.2 },
      { freq: 783.99, time: now + 0.16, dur: 0.35 },
    ]

    const masterGain = ctx.createGain()
    masterGain.gain.setValueAtTime(0.0001, now)
    masterGain.gain.exponentialRampToValueAtTime(volume * 0.7, now + 0.05)
    masterGain.gain.exponentialRampToValueAtTime(0.0001, now + 0.55)
    masterGain.connect(ctx.destination)

    notes.forEach(({ freq, time, dur }) => {
      const osc = ctx.createOscillator()
      const noteGain = ctx.createGain()

      osc.type = 'sine'
      osc.frequency.setValueAtTime(freq, time)

      noteGain.gain.setValueAtTime(0.0001, time)
      noteGain.gain.exponentialRampToValueAtTime(0.8, time + 0.02)
      noteGain.gain.exponentialRampToValueAtTime(0.0001, time + dur)

      osc.connect(noteGain)
      noteGain.connect(masterGain)

      osc.start(time)
      osc.stop(time + dur + 0.05)
    })
  } catch {
    // Gracefully ignore browser audio policy restrictions
  }
}

/**
 * Play a clear, audible micro-click / toggle sound
 */
export function playSoftClick(pitch = 600) {
  if (!isSoundEnabled()) return

  try {
    const ctx = getAudioContext()
    if (!ctx) return

    if (ctx.state === 'suspended') {
      ctx.resume()
    }

    const volume = getSoundVolume()
    const now = ctx.currentTime
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()

    osc.type = 'triangle'
    osc.frequency.setValueAtTime(pitch, now)
    osc.frequency.exponentialRampToValueAtTime(pitch * 0.75, now + 0.06)

    gain.gain.setValueAtTime(0.0001, now)
    gain.gain.exponentialRampToValueAtTime(volume * 0.6, now + 0.008)
    gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.07)

    osc.connect(gain)
    gain.connect(ctx.destination)

    osc.start(now)
    osc.stop(now + 0.08)
  } catch {
    // Gracefully ignore
  }
}

/**
 * Play a distinctive sound when toggling Privacy / Discreet mode
 */
export function playPrivacyToggleSound(isDiscreet: boolean) {
  if (!isSoundEnabled()) return

  try {
    const ctx = getAudioContext()
    if (!ctx) return

    if (ctx.state === 'suspended') {
      ctx.resume()
    }

    const volume = getSoundVolume()
    const now = ctx.currentTime
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()

    osc.type = 'sine'

    if (isDiscreet) {
      // Descending tone: Masking / Lock effect (660Hz -> 440Hz)
      osc.frequency.setValueAtTime(659.25, now)
      osc.frequency.exponentialRampToValueAtTime(440.0, now + 0.12)
    } else {
      // Ascending tone: Reveal / Unlock effect (440Hz -> 660Hz)
      osc.frequency.setValueAtTime(440.0, now)
      osc.frequency.exponentialRampToValueAtTime(659.25, now + 0.12)
    }

    gain.gain.setValueAtTime(0.0001, now)
    gain.gain.exponentialRampToValueAtTime(volume * 0.65, now + 0.02)
    gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.18)

    osc.connect(gain)
    gain.connect(ctx.destination)

    osc.start(now)
    osc.stop(now + 0.2)
  } catch {
    // Gracefully ignore
  }
}
