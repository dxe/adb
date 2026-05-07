/**
 * IndexedDB storage for activist list data and sync metadata.
 * Provides caching and incremental sync capabilities.
 *
 * Why IndexedDB instead of localStorage:
 * - Data size: ~3MB as of January 2026, approaching localStorage's 5-10MB limit
 * - Async operations: Avoid blocking main thread with large JSON parse/stringify
 * - Structured storage: Store by ID with built-in indexing
 */

export interface StoredActivist {
  id: number
  name: string
  email: boolean
  phone: boolean
  lastUpdated: number // Unix timestamp in milliseconds
  lastEventDate: number // Unix timestamp in milliseconds, 0 if no events
}

interface SyncMetadata {
  lastSyncTime: string // ISO 8601 timestamp
}

const DB_VERSION = 2
const STORE_NAME = 'activists'
const METADATA_STORE = 'metadata'

export class ActivistStorage {
  private dbPromise: Promise<IDBDatabase> | null = null
  private readonly dbName: string

  constructor(chapterId: number) {
    this.dbName = `adb-chapter-${chapterId}`
  }

  /**
   * Initialize the IndexedDB database with required object stores.
   */
  private openDB(): Promise<IDBDatabase> {
    if (!this.dbPromise) {
      this.dbPromise = new Promise((resolve, reject) => {
        const request = indexedDB.open(this.dbName, DB_VERSION)

        request.onerror = () => reject(request.error)
        request.onsuccess = () => {
          const db = request.result
          // Close connection immediately if another tab needs to upgrade
          db.onversionchange = () => {
            db.close()
            this.dbPromise = null // Reset cache so subsequent calls reopen
          }
          resolve(db)
        }

        // Reject if upgrade is blocked by another tab with an open connection
        request.onblocked = () => {
          this.dbPromise = null // Reset cache to allow retry
          reject(
            new Error(
              `Database upgrade to version ${DB_VERSION} blocked by another tab. Please close other tabs and refresh.`,
            ),
          )
        }

        request.onupgradeneeded = (event) => {
          const db = (event.target as IDBOpenDBRequest).result
          const oldVersion = event.oldVersion
          const transaction = (event.target as IDBOpenDBRequest).transaction!

          // Migration from v1 to v2: Clear cache AND metadata to force full sync
          if (oldVersion === 1) {
            if (db.objectStoreNames.contains(STORE_NAME)) {
              const activistStore = transaction.objectStore(STORE_NAME)
              activistStore.clear()
            }
            if (db.objectStoreNames.contains(METADATA_STORE)) {
              const metadataStore = transaction.objectStore(METADATA_STORE)
              metadataStore.clear()
            }
          }

          // Create activists store with id as key
          if (!db.objectStoreNames.contains(STORE_NAME)) {
            db.createObjectStore(STORE_NAME, { keyPath: 'id' })
          }

          // Create metadata store for sync timestamps
          if (!db.objectStoreNames.contains(METADATA_STORE)) {
            db.createObjectStore(METADATA_STORE)
          }
        }
      })
    }
    return this.dbPromise
  }

  /**
   * Get the last sync timestamp from IndexedDB.
   */
  async getLastSyncTime(): Promise<string | null> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const transaction = db.transaction([METADATA_STORE], 'readonly')
      const store = transaction.objectStore(METADATA_STORE)
      const request = store.get('lastActivistSync')

      request.onsuccess = () => {
        const metadata = request.result as SyncMetadata | undefined
        resolve(metadata?.lastSyncTime ?? null)
      }
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Update or delete the last sync timestamp in IndexedDB.
   * Pass null to delete the timestamp (forcing a full sync on next load).
   */
  async setLastSyncTime(timestamp: string | null): Promise<void> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const transaction = db.transaction([METADATA_STORE], 'readwrite')
      const store = transaction.objectStore(METADATA_STORE)

      const request =
        timestamp === null
          ? store.delete('lastActivistSync')
          : store.put(
              { lastSyncTime: timestamp } satisfies SyncMetadata,
              'lastActivistSync',
            )

      request.onsuccess = () => resolve()
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Gets all activists from IndexedDB.
   */
  async getAllActivists(): Promise<StoredActivist[]> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const transaction = db.transaction([STORE_NAME], 'readonly')
      const store = transaction.objectStore(STORE_NAME)
      const request = store.getAll()

      request.onsuccess = () => resolve(request.result)
      request.onerror = () => reject(request.error)
    })
  }

  /**
   * Store or update activists in IndexedDB.
   * Uses upsert semantics - adds new activists and updates existing ones.
   * @throws Error if quota exceeded or transaction fails
   */
  async saveActivists(activists: StoredActivist[]): Promise<void> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const transaction = db.transaction([STORE_NAME], 'readwrite')
      const store = transaction.objectStore(STORE_NAME)

      // Use put() for upsert behavior (insert or update)
      for (const activist of activists) {
        store.put(activist)
      }

      transaction.oncomplete = () => resolve()
      transaction.onerror = () => {
        reject(transaction.error)
      }
    })
  }

  /**
   * Delete activists by their IDs from IndexedDB.
   * Used for syncing deletions/hidden activists.
   */
  async deleteActivistsByIds(ids: number[]): Promise<void> {
    if (ids.length === 0) return

    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const transaction = db.transaction([STORE_NAME], 'readwrite')
      const store = transaction.objectStore(STORE_NAME)

      for (const id of ids) {
        store.delete(id)
      }

      transaction.oncomplete = () => resolve()
      transaction.onerror = () => reject(transaction.error)
    })
  }

  /**
   * Clear all activist data from IndexedDB.
   * Useful for forcing a full refresh.
   */
  async clearAllActivists(): Promise<void> {
    const db = await this.openDB()
    return new Promise((resolve, reject) => {
      const transaction = db.transaction([STORE_NAME], 'readwrite')
      const store = transaction.objectStore(STORE_NAME)
      const request = store.clear()

      request.onsuccess = () => resolve()
      request.onerror = () => reject(request.error)
    })
  }
}

const ENABLE_ACTIVIST_REGISTRY = false

/**
 * Check if IndexedDB is available and accessible.
 * Returns false in environments where IndexedDB is blocked (e.g., iOS lockdown mode).
 */
function isIndexedDBAvailable(): boolean {
  try {
    return typeof indexedDB !== 'undefined' && indexedDB !== null
  } catch {
    return false
  }
}

const storageByChapter = new Map<number, ActivistStorage>()

export function getActivistStorage(
  chapterId: number,
): ActivistStorage | undefined {
  if (!ENABLE_ACTIVIST_REGISTRY) return undefined
  if (!isIndexedDBAvailable()) return undefined
  const existing = storageByChapter.get(chapterId)
  if (existing) return existing
  const storage = new ActivistStorage(chapterId)
  storageByChapter.set(chapterId, storage)
  return storage
}
