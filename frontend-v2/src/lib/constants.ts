/**
 * Shared constants used across the application.
 */

/**
 * Chapter ID for SF Bay Area chapter.
 * Different values for production vs development environments.
 */
export const SF_BAY_CHAPTER_ID = process.env.NODE_ENV === 'production' ? 47 : 1
