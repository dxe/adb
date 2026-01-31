/**
 * IndexedDB storage for activist list data and sync metadata.
 * Provides caching and incremental sync capabilities.
 *
 * Why IndexedDB instead of localStorage:
 * - Data size: ~3MB currently, approaching localStorage's 5-10MB limit
 * - Async operations: Avoid blocking main thread with large JSON parse/stringify
 * - Structured storage: Store by ID with built-in indexing
 */

const DB_NAME = 'activist-registry'
const DB_VERSION = 2 // Incremented to change keyPath from 'name' to 'id'
const STORE_NAME = 'activists'
const METADATA_STORE = 'metadata'

export type StoredActivist = {
  id: number
  name: string
  email?: string
  phone?: string
}

type SyncMetadata = {
  lastSyncTime: string // ISO 8601 timestamp
}

/**
 * Initialize the IndexedDB database with required object stores.
 */
function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)

    request.onerror = () => reject(request.error)
    request.onsuccess = () => resolve(request.result)

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result
      const oldVersion = event.oldVersion

      // Migrate from version 1 to 2: change keyPath from 'name' to 'id'
      if (oldVersion < 2) {
        // Delete old store if it exists (name-keyed)
        if (db.objectStoreNames.contains(STORE_NAME)) {
          db.deleteObjectStore(STORE_NAME)
        }
        // Create new store with id as key
        db.createObjectStore(STORE_NAME, { keyPath: 'id' })
      }

      // Create activists store with id as key (for fresh installs)
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

/**
 * Get the last sync timestamp from IndexedDB.
 */
export async function getLastSyncTime(): Promise<string | null> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([METADATA_STORE], 'readonly')
    const store = transaction.objectStore(METADATA_STORE)
    const request = store.get('lastSync')

    request.onsuccess = () => {
      const metadata = request.result as SyncMetadata | undefined
      resolve(metadata?.lastSyncTime ?? null)
    }
    request.onerror = () => reject(request.error)
  })
}

/**
 * Update the last sync timestamp in IndexedDB.
 */
export async function setLastSyncTime(timestamp: string): Promise<void> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([METADATA_STORE], 'readwrite')
    const store = transaction.objectStore(METADATA_STORE)
    const metadata: SyncMetadata = { lastSyncTime: timestamp }
    const request = store.put(metadata, 'lastSync')

    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

/**
 * Get all activists from IndexedDB.
 */
export async function getAllActivists(): Promise<StoredActivist[]> {
  const db = await openDB()
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
 */
export async function saveActivists(
  activists: StoredActivist[],
): Promise<void> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORE_NAME], 'readwrite')
    const store = transaction.objectStore(STORE_NAME)

    // Use put() for upsert behavior (insert or update)
    for (const activist of activists) {
      store.put(activist)
    }

    transaction.oncomplete = () => resolve()
    transaction.onerror = () => reject(transaction.error)
  })
}

/**
 * Delete activists by their IDs from IndexedDB.
 * Used for syncing deletions/hidden activists.
 */
export async function deleteActivistsByIds(ids: number[]): Promise<void> {
  if (ids.length === 0) return

  const db = await openDB()
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
export async function clearAllActivists(): Promise<void> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([STORE_NAME], 'readwrite')
    const store = transaction.objectStore(STORE_NAME)
    const request = store.clear()

    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}
