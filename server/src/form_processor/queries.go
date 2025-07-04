package form_processor

import "github.com/dxe/adb/model"

/* Common queries */
const countActivistsForEmailQuery = "SELECT count(id) AS amount FROM activists WHERE hidden = 0 and email = ? and chapter_id = " + model.SFBayChapterIdStr
