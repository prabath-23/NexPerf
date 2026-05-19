import { mkdirSync, writeFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const output = resolve(scriptDir, '..', 'public', 'brand', 'favicon.ico')
const sizes = [16, 32]

mkdirSync(dirname(output), { recursive: true })
writeFileSync(output, createIco(sizes))
console.log(`Generated ${output}`)

function createIco(iconSizes) {
  const images = iconSizes.map((size) => ({ size, dib: createDib(size) }))
  const headerSize = 6 + images.length * 16
  const totalSize = headerSize + images.reduce((sum, image) => sum + image.dib.length, 0)
  const ico = Buffer.alloc(totalSize)
  let offset = 0

  ico.writeUInt16LE(0, offset)
  offset += 2
  ico.writeUInt16LE(1, offset)
  offset += 2
  ico.writeUInt16LE(images.length, offset)
  offset += 2

  let imageOffset = headerSize
  for (const image of images) {
    ico.writeUInt8(image.size === 256 ? 0 : image.size, offset)
    ico.writeUInt8(image.size === 256 ? 0 : image.size, offset + 1)
    ico.writeUInt8(0, offset + 2)
    ico.writeUInt8(0, offset + 3)
    ico.writeUInt16LE(1, offset + 4)
    ico.writeUInt16LE(32, offset + 6)
    ico.writeUInt32LE(image.dib.length, offset + 8)
    ico.writeUInt32LE(imageOffset, offset + 12)
    offset += 16
    image.dib.copy(ico, imageOffset)
    imageOffset += image.dib.length
  }

  return ico
}

function createDib(size) {
  const bitmapHeaderSize = 40
  const pixelBytes = size * size * 4
  const maskStride = Math.ceil(size / 32) * 4
  const maskBytes = maskStride * size
  const dib = Buffer.alloc(bitmapHeaderSize + pixelBytes + maskBytes)

  dib.writeUInt32LE(bitmapHeaderSize, 0)
  dib.writeInt32LE(size, 4)
  dib.writeInt32LE(size * 2, 8)
  dib.writeUInt16LE(1, 12)
  dib.writeUInt16LE(32, 14)
  dib.writeUInt32LE(0, 16)
  dib.writeUInt32LE(pixelBytes + maskBytes, 20)

  const signalSegments = [
    [[0.25, 0.74], [0.25, 0.26]],
    [[0.25, 0.26], [0.5, 0.74]],
    [[0.5, 0.74], [0.5, 0.26]],
    [[0.5, 0.26], [0.75, 0.74]],
    [[0.75, 0.74], [0.75, 0.26]],
    [[0.3, 0.58], [0.42, 0.58]],
    [[0.42, 0.58], [0.49, 0.48]],
    [[0.49, 0.48], [0.59, 0.62]],
    [[0.59, 0.62], [0.66, 0.38]],
    [[0.66, 0.38], [0.73, 0.5]],
    [[0.73, 0.5], [0.84, 0.5]]
  ]

  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const nx = (x + 0.5) / size
      const ny = (y + 0.5) / size
      const pixelOffset = bitmapHeaderSize + ((size - 1 - y) * size + x) * 4
      const bg = roundedRect(nx, ny, 0.18)
      const signal = nearSegments(nx, ny, signalSegments, size <= 16 ? 0.085 : 0.052)
      const pulse = nearSegments(nx, ny, signalSegments.slice(5), size <= 16 ? 0.06 : 0.036)

      if (!bg) continue

      const bgShade = 16 + Math.round(26 * (nx + ny) / 2)
      setPixel(dib, pixelOffset, bgShade, bgShade + 8, bgShade + 18, 255)

      if (signal) {
        const t = Math.min(1, Math.max(0, (nx + 1 - ny) / 1.6))
        const color = pulse ? [19, 184, 166] : mix([49, 87, 213], [19, 184, 166], t)
        setPixel(dib, pixelOffset, color[0], color[1], color[2], 255)
      }
    }
  }

  return dib
}

function roundedRect(x, y, radius) {
  const dx = Math.max(Math.abs(x - 0.5) - 0.5 + radius, 0)
  const dy = Math.max(Math.abs(y - 0.5) - 0.5 + radius, 0)
  return dx * dx + dy * dy <= radius * radius
}

function nearSegments(x, y, segments, threshold) {
  for (const segment of segments) {
    const [a, b] = segment
    if (distanceToSegment(x, y, a[0], a[1], b[0], b[1]) <= threshold) return true
  }
  return false
}

function distanceToSegment(px, py, ax, ay, bx, by) {
  const dx = bx - ax
  const dy = by - ay
  const length = dx * dx + dy * dy || 1
  const t = Math.max(0, Math.min(1, ((px - ax) * dx + (py - ay) * dy) / length))
  const x = ax + t * dx
  const y = ay + t * dy
  return Math.hypot(px - x, py - y)
}

function mix(a, b, t) {
  return a.map((value, index) => Math.round(value + (b[index] - value) * t))
}

function setPixel(buffer, offset, r, g, b, a) {
  buffer[offset] = clamp(b)
  buffer[offset + 1] = clamp(g)
  buffer[offset + 2] = clamp(r)
  buffer[offset + 3] = clamp(a)
}

function clamp(value) {
  return Math.max(0, Math.min(255, Math.round(value)))
}
