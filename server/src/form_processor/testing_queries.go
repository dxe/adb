package form_processor

import "github.com/dxe/adb/model"

/* Common queries */
const insertActivistQuery = `
INSERT INTO activists (id, email, name, chapter_id) VALUES (NULL, "email1", ?, ` + model.SFBayChapterIdStr + `);
`

const getActivistsQuery = `SELECT id FROM activists;`

/* Form application queries */
const insertIntoFormApplicationQuery = `
INSERT INTO form_application (
  id,
  email,
  name,
  phone,
  address,
  city,
  zip,
  birthday,
  pronouns,
  application_type,
  agree_circle,
  agree_mpp,
  circle_interest,
  wg_interest,
  committee_interest,
  referral_friends,
  referral_apply,
  referral_outlet,
  contact_method,
  processed
) VALUES (
  NULL,
  "email1",
  "name1",
  "phone1",
  "address1",
  "city1",
  "zip1",
  "birthday1",
  "pronouns1",
  "application_type1",
  "agree_circle1",
  "agree_mpp1",
  "circle_interest1",
  "wg_interest1",
  "committee_interest1",
  "referral_friends1",
  "referral_apply1",
  "referral_outlet1",
  "contact_method1",
  false
);
`

/* Form interest queries */
const insertIntoFormInterestQuery = `
INSERT INTO form_interest (
  id,
  form,
  email,
  name,
  phone,
  zip,
  referral_friends,
  referral_apply,
  referral_outlet,
  comments,
  interests
) VALUES (
  NULL,
  "form1",
  "email1",
  "name1",
  "phone1",
  "zip1",
  "referral_friends1",
  "referral_apply1",
  "referral_outlet1",
  "comments1",
  "interests1"
);
`
